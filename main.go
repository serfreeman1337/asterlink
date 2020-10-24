package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	"github.com/serfreeman1337/asterlink/connect/b24"
	"github.com/serfreeman1337/asterlink/connect/suitecrm"

	"gopkg.in/yaml.v2"

	"github.com/ivahaev/amigo"
	log "github.com/sirupsen/logrus"
)

type numForm struct {
	R    *regexp.Regexp
	Expr string
	Repl string
}

type config struct {
	LogLevel log.Level `yaml:"log_level"`
	Ami      struct {
		Host string
		Port string
		User string
		Pass string
	}
	DP struct {
		In   []string `yaml:"incoming_context"`
		Out  []string `yaml:"outgoing_context"`
		Ext  []string `yaml:"ext_context"`
		Dial string   `yaml:"dial_context"`
		Vote string   `yaml:"vote_ivr"`
	} `yaml:"dialplan"`
	Form struct {
		hasCid  bool
		hasDial bool
		hasDid  bool
		Cid     []numForm `yaml:"cid_format"`
		Dial    []numForm `yaml:"dial_format"`
		LimDID  []string  `yaml:"dids"`
		DID     map[string]bool
	} `yaml:"pbx"`
	B24      b24.Config      `yaml:"bitrix24"`
	SuiteCRM suitecrm.Config `yaml:"suitecrm"`
}

func (cfg *config) getConf() {
	yamlFile, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Ami.Host == "" || cfg.Ami.User == "" || cfg.Ami.Pass == "" {
		log.Fatal("AMI settings are missing from config file")
	} else if len(cfg.DP.In) == 0 || len(cfg.DP.Out) == 0 || len(cfg.DP.Ext) == 0 || cfg.DP.Dial == "" {
		log.Fatal("DialPlan configuration are missing from config file")
	}

	if cfg.Ami.Port == "" {
		cfg.Ami.Port = "5038"
	}
	if cfg.LogLevel == log.PanicLevel {
		cfg.LogLevel = log.TraceLevel
	}

	if len(cfg.Form.Cid) != 0 {
		for i, v := range cfg.Form.Cid {
			if cfg.Form.Cid[i].R, err = regexp.Compile(v.Expr); err != nil {
				log.Fatal(err)
			}
		}
		cfg.Form.hasCid = true
	}
	if len(cfg.Form.Dial) != 0 {
		for i, v := range cfg.Form.Dial {
			if cfg.Form.Dial[i].R, err = regexp.Compile(v.Expr); err != nil {
				log.Fatal(err)
			}
		}
		cfg.Form.hasDial = true
	}
	if len(cfg.Form.LimDID) != 0 {
		cfg.Form.hasDid = true
		cfg.Form.DID = make(map[string]bool)

		for _, d := range cfg.Form.LimDID {
			cfg.Form.DID[d] = true
		}
	}

	if cfg.B24.URL != "" {
		if len(cfg.B24.FindForm) != 0 {
			for i, v := range cfg.B24.FindForm {
				if cfg.B24.FindForm[i].R, err = regexp.Compile(v.Expr); err != nil {
					log.Fatal(err)
				}
			}
			cfg.B24.HasFindForm = true
		}
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	log.Info("AsterLink")

	var cfg config
	cfg.getConf()

	log.WithField("level", cfg.LogLevel).Info("Setting log level")
	log.SetLevel(cfg.LogLevel)

	cdr := make(map[string]*connect.Call)
	rec := make(map[string]string)

	settings := &amigo.Settings{Host: cfg.Ami.Host, Username: cfg.Ami.User, Password: cfg.Ami.Pass, Port: cfg.Ami.Port}
	ami := amigo.New(settings)

	useRec := func(c *connect.Call, uID string) {
		if v, ok := rec[uID]; ok {
			c.Rec = v
			delete(rec, uID)

			c.Log.WithField("rec", c.Rec).Debug("Set recording file")
		}
	}

	formatNum := func(cID string, isCid bool) (string, bool) {
		var hasForm bool
		var form []numForm
		if isCid {
			hasForm = cfg.Form.hasCid
			form = cfg.Form.Cid
		} else {
			hasForm = cfg.Form.hasDial
			form = cfg.Form.Dial
		}

		if !hasForm {
			return cID, true
		}

		for _, f := range form {
			if f.R.MatchString(cID) {
				return f.R.ReplaceAllString(cID, f.Repl), true
			}
		}

		return "", false
	}

	isContext := func(context string, of []string) bool {
		for _, v := range of {
			if context == v {
				return true
			}
		}

		return false
	}

	originate := func(ext string, dest string, oID string) {
		cLog := log.WithFields(log.Fields{"ext": ext, "dest": dest, "oID": oID})
		num, ok := formatNum(dest, false)

		if !ok {
			cLog.WithField("num", num).Error("Invalid number to originate call")
			return
		}

		res, err := ami.Action(map[string]string{
			"Action":   "Originate",
			"Channel":  "Local/" + ext + "@" + cfg.DP.Dial,
			"Exten":    num,
			"Context":  cfg.DP.Dial,
			"Priority": "1",
			"Variable": "SF_CONNECTOR=" + oID + "|" + dest, // keep track of call record and original caller id
			"CallerID": ext,
			"Async":    "true",
			"Codecs":   "alaw,ulaw", // TODO: config?
		})
		if err != nil {
			cLog.Error(err)
		} else if res["Response"] != "Success" {
			cLog.Error(res["Message"])
		} else {
			cLog.Debug("New originate request")
		}
	}

	rechan := regexp.MustCompile(`.*\/(\d+)`)
	var revote *regexp.Regexp

	var connector connect.Connecter

	if cfg.B24.URL != "" {
		log.Info("Using Bitrix24 Connector")
		connector = b24.NewB24Connector(&cfg.B24, originate)
	} else if cfg.SuiteCRM.URL != "" {
		log.Info("Using SuiteCRM Connector")
		connector = suitecrm.NewSuiteCRMConnector(&cfg.SuiteCRM, originate)
	} else {
		log.Warn("No connector selected")
		connector = connect.NewDummyConnector()
	}

	connector.Init()

	ami.On("connect", func(message string) {
		log.WithField("msg", message).Info("Established connection with AMI")

		fexpr := func(r []string) string {
			var s string
			if len(r) == 1 {
				s += r[0]
			} else {
				s += "("

				for _, v := range r {
					s += v + "|"
				}

				s = s[:len(s)-1]
				s += ")"
			}

			return s
		}

		// posix regexp
		filter := "(Event: Newchannel.*[[:cntrl:]]Context: " + fexpr(cfg.DP.In) + "[[:cntrl:]]"
		filter += "|Event: (DialBegin|DialEnd).*[[:cntrl:]]Context: " + fexpr(append(cfg.DP.Ext, cfg.DP.Out...)) + "[[:cntrl:]]"
		filter += "|Event: Newexten.*[[:cntrl:]]"

		if cfg.DP.Vote == "" {
			filter += "Application: MixMonitor"
		} else {
			filter += "(Application: MixMonitor|Context: " + cfg.DP.Vote + "[[:cntrl:]].*AppData: CDR\\(userfield\\)=\".*rate.*\")"
			filter += "|Variable: IVR_CONTEXT.*Value: " + cfg.DP.Vote
			revote = regexp.MustCompile(`CDR\(userfield\)="rate:(\d+)"`)
		}

		filter += "|Event: (Hangup|BlindTransfer)|Variable: SF_CONNECTOR)"

		// let the Asterisk filter events for us
		res, err := ami.Action(map[string]string{"Action": "Filter", "Operation": "Add", "Filter": filter})
		if err != nil {
			log.WithField("filter", filter).Fatal(err)
		} else if res["Response"] != "Success" {
			log.WithField("filter", filter).Fatal(res["Message"])
		}
	})

	ami.On("error", func(message string) {
		log.Error(message)
	})

	// incoming call
	ami.RegisterHandler("Newchannel", func(e map[string]string) {
		if /*_, ok := IncomingContext[e["Context"]]; !ok || */ e["Exten"] == "s" { // TODO: review this 's'!
			return
		}

		if _, ok := cdr[e["Linkedid"]]; ok {
			log.WithField("lid", e["Linkedid"]).Warn("Already tracked")

			return
		}

		// DID Limit
		if cfg.Form.hasDid {
			if _, ok := cfg.Form.DID[e["Exten"]]; !ok {
				log.WithFields(log.Fields{"lid": e["Linkedid"], "did": e["Exten"]}).Debug("Skip incoming call for DID")
				return
			}
		}

		// skip anonymous calls ?
		if e["CallerIDNum"] == "<unknown>" {
			log.WithFields(log.Fields{"lid": e["Linkedid"]}).Warn("Anonymous CallerID")
			return
		}

		cID, ok := formatNum(e["CallerIDNum"], true)
		if !ok {
			log.WithFields(log.Fields{"lid": e["Linkedid"], "cid": e["CallerIDNum"]}).Warn("Unknown incoming CallerID")
			return
		}

		// register incoming call
		cdr[e["Linkedid"]] = &connect.Call{
			LID:      e["Linkedid"],
			Dir:      connect.In,
			CID:      cID,
			DID:      e["Exten"],
			TimeCall: time.Now(),
			Ch:       e["Channel"],
			Log:      log.WithField("lid", e["Linkedid"]),
		}
		cdr[e["Linkedid"]].Log.Debug("New incoming call")

		go connector.Start(cdr[e["Linkedid"]])
	})

	// operator dial
	ami.RegisterHandler("DialBegin", func(e map[string]string) {
		c, ok := cdr[e["Linkedid"]]

		if !ok {
			// TODO: review
			if !isContext(e["Context"], cfg.DP.Out) {
				return
			}

			cID, ok := formatNum(e["DestCallerIDNum"], true)
			if !ok {
				log.WithFields(log.Fields{"lid": e["Linkedid"], "cid": e["DestCallerIDNum"]}).Warn("Unknown outgoing CallerID")
				return
			}

			rext := rechan.FindStringSubmatch(e["Channel"])
			if len(rext) == 0 {
				return
			}

			// register outbound call
			c = &connect.Call{
				LID:      e["Linkedid"],
				Dir:      connect.Out,
				CID:      cID,
				DID:      e["CallerIDNum"],
				TimeCall: time.Now(),
				TimeDial: time.Now(),
				Ch:       e["Channel"],
				ChDest:   e["DestChannel"],
				Log:      log.WithField("lid", e["Linkedid"]),
				Ext:      rext[1],
			}
			cdr[e["Linkedid"]] = c

			c.Log.Debug("New outgoing call")

			go func() {
				connector.Start(c)
				connector.Dial(c, rext[1])
			}()

			return
		} else if c.O { // Originated call
			if isContext(e["Context"], cfg.DP.Ext) {
				c.Ch = e["DestChannel"]
				c.Ext = e["DestCallerIDNum"]

				c.Log.WithField("ext", c.Ext).Debug("From")
				useRec(c, e["Uniqueid"])

				go connector.Dial(c, c.Ext)
			} else if isContext(e["Context"], cfg.DP.Out) {
				// TODO: do something about originated CallerID
				c.ChDest = e["DestChannel"]
				c.TimeDial = time.Now()
				c.DID = e["DestConnectedLineNum"]

				c.Log.WithField("did", c.DID).Debug("Via")
			}

			return
		}

		/*if _, ok := ExtContext[e["Context"]]; !ok {
			return
		}*/

		if c.TimeDial.IsZero() {
			c.TimeDial = time.Now()
		}

		c.Log.WithField("ext", e["DestCallerIDNum"]).Debug("Dial")
		go connector.Dial(c, e["DestCallerIDNum"])
	})

	// operator answer
	ami.RegisterHandler("DialEnd", func(e map[string]string) {
		c, ok := cdr[e["Linkedid"]]

		if !ok {
			return
		}

		if e["DialStatus"] != "ANSWER" {
			return
		}

		switch c.Dir {
		case connect.In:
			/*if _, ok := ExtContext[e["Context"]]; !ok {
				return
			}*/

			c.TimeAnswer = time.Now()
			c.ChDest = e["DestChannel"]
			c.Ext = e["DestCallerIDNum"]

			useRec(c, e["Uniqueid"])

			c.Log.WithField("ext", c.Ext).Debug("Answer")
			go connector.Answer(c, c.Ext)
			break
		case connect.Out:
			if c.O && !isContext(e["Context"], cfg.DP.Out) { // Originated call
				return
			}

			// TODO: the heck am i doing ?
			if isContext(e["Context"], cfg.DP.Ext) && c.ChDest == e["Channel"] {
				// BlindTransfer hack ?
				rext := rechan.FindStringSubmatch(e["DestChannel"])
				if len(rext) == 0 {
					return
				}

				c.Ext = rext[1]
				c.Log.WithField("ext", c.Ext).Debug("Answer 2")
				useRec(c, e["Uniqueid"])
				go connector.Answer(c, c.Ext)

				return
			}

			rext := rechan.FindStringSubmatch(e["Channel"])
			if len(rext) == 0 {
				return
			}

			c.TimeAnswer = time.Now()
			c.Ext = rext[1]

			c.Log.WithField("ext", c.Ext).Debug("Dial 2")
			useRec(c, e["Uniqueid"])

			go connector.Answer(c, rext[1])

			break
		}
	})

	// call recording and vote
	ami.RegisterHandler("Newexten", func(e map[string]string) {
		if e["Application"] == "MixMonitor" { // recording
			rec[e["Uniqueid"]] = strings.Split(e["AppData"], ",")[0]
			log.WithFields(log.Fields{"lid": e["Linkedid"], "uid": e["Uniqueid"], "rec": rec[e["Uniqueid"]]}).Debug("MixMonitor")
		} else { // vote
			c, ok := cdr[e["Linkedid"]]
			if !ok {
				return
			}
			appData := e["AppData"][:len(e["AppData"])-1] // 3 IQ WHITESPACE TRIM
			c.Vote = revote.ReplaceAllString(appData, "$1")
			c.Log.WithField("vote", c.Vote).Debug("Call voted")
		}
	})

	// TODO: AttendedTransfer
	ami.RegisterHandler("BlindTransfer", func(e map[string]string) {
		// what ?
		lID, ok := e["TransfererLinkedid"]
		if !ok {
			lID, ok = e["Linkedid"]
			if !ok {
				log.Warn("Unknown BlindTransfer")
				log.Warn(e)

				return
			}
		}

		c, ok := cdr[lID]
		if !ok {
			return
		}

		// forget original operator
		if c.Dir == connect.In {
			c.ChDest = ""
		} else if c.Dir == connect.Out {
			c.Ch = ""
		}

		c.Log.WithField("ext", c.Ext).Debug("BlindTransfer")
		go connector.StopDial(c, c.Ext)
	})

	// call finish
	ami.RegisterHandler("Hangup", func(e map[string]string) {
		_, ok := rec[e["Uniqueid"]]

		if ok {
			log.WithFields(log.Fields{"lid": e["Linkedid"], "uid": e["Uniqueid"]}).Debug("Clear recording")
			delete(rec, e["Uniqueid"])
		}

		c, ok := cdr[e["Linkedid"]]

		if !ok {
			return
		}

		if e["Channel"] == c.Ch || e["Channel"] == c.ChDest {
			// call marked for vote, keep tracking after operator hangup
			if c.Vote == "-" && !c.TimeAnswer.IsZero() && e["Channel"] == c.ChDest {
				go connector.StopDial(c, c.Ext)

				c.ChDest = ""
				c.Log.Debug("Wait for vote")

				return
			}

			c.Log.Debug("Call finished")
			go connector.End(c, e["Cause"])

			delete(cdr, e["Linkedid"])
		} else {
			if e["Context"] == cfg.DP.Dial && e["ChannelState"] == "5" {
				c.Log.WithField("ext", e["CallerIDNum"]).Debug("Dial stop")
				go connector.StopDial(c, e["CallerIDNum"])
			}
		}
	})

	// originated call
	ami.RegisterHandler("VarSet", func(e map[string]string) {
		c, ok := cdr[e["Linkedid"]]
		if !ok {
			// VarSet for originated call
			if e["Variable"] != "SF_CONNECTOR" || e["Exten"] == "failed" {
				return
			}

			// split tracker variable (see Originate request)
			r := strings.Split(e["Value"], "|")

			cdr[e["Linkedid"]] = &connect.Call{
				LID:      e["Linkedid"],
				CID:      r[1],
				TimeCall: time.Now(),
				TimeDial: time.Now(),
				Dir:      connect.Out,
				Ch:       e["Channel"],
				O:        true,
				Log:      log.WithField("lid", e["Linkedid"]),
			}

			log.WithField("oid", r[0]).Debug("New originated call")
			connector.OrigStart(cdr[e["Linkedid"]], r[0])

			return
		}

		if cfg.DP.Vote == "" || e["Value"] != cfg.DP.Vote {
			return
		}

		// Mark call for vote
		c.Vote = "-"
	})

	ami.Connect()

	r := make(chan bool)
	<-r
}
