package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/serfreeman1337/asterlink/connect"

	"gopkg.in/yaml.v2"

	"github.com/ivahaev/amigo"
	log "github.com/sirupsen/logrus"
)

func isContext(context string, of []string) bool {
	for _, v := range of {
		if context == v {
			return true
		}
	}

	return false
}

type numForm struct {
	R    *regexp.Regexp
	Expr string
	Repl string
}

type conf struct {
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
	B24 struct {
		Addr     string `yaml:"webhook_endpoint_addr"`
		URL      string `yaml:"webhook_url"`
		Token    string `yaml:"webhook_originate_token"`
		hasSForm bool
		SForm    []numForm `yaml:"search_format"`
	} `yaml:"bitrix24"`
}

func (c *conf) getConf() {
	yamlFile, err := ioutil.ReadFile("conf.yml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatal(err)
	}

	if c.Ami.Host == "" || c.Ami.User == "" || c.Ami.Pass == "" {
		log.Fatal("AMI settings are missing from config file")
	} else if len(c.DP.In) == 0 || len(c.DP.Out) == 0 || len(c.DP.Ext) == 0 || c.DP.Dial == "" {
		log.Fatal("DialPlan configuration are missing from config file")
	}

	if c.Ami.Port == "" {
		c.Ami.Port = "5038"
	}
	if c.LogLevel == log.PanicLevel {
		c.LogLevel = log.TraceLevel
	}

	if len(c.Form.Cid) != 0 {
		for i, v := range c.Form.Cid {
			if c.Form.Cid[i].R, err = regexp.Compile(v.Expr); err != nil {
				log.Fatal(err)
			}
		}
		c.Form.hasCid = true
	}
	if len(c.Form.Dial) != 0 {
		for i, v := range c.Form.Dial {
			if c.Form.Dial[i].R, err = regexp.Compile(v.Expr); err != nil {
				log.Fatal(err)
			}
		}
		c.Form.hasDial = true
	}
	if len(c.Form.LimDID) != 0 {
		c.Form.hasDid = true
		c.Form.DID = make(map[string]bool)

		for _, d := range c.Form.LimDID {
			c.Form.DID[d] = true
		}
	}

	if c.B24.URL != "" {
		if len(c.B24.SForm) != 0 {
			for i, v := range c.B24.SForm {
				if c.B24.SForm[i].R, err = regexp.Compile(v.Expr); err != nil {
					log.Fatal(err)
				}
			}
			c.B24.hasSForm = true
		}
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	log.Info("AsterLink v. 0.2.0-dev")
}

func main() {
	var cfg conf
	cfg.getConf()

	log.WithField("level", cfg.LogLevel).Info("Setting log level")
	log.SetLevel(cfg.LogLevel)

	cdr := make(map[string]*connect.Call)
	rec := make(map[string]string)

	rechan := regexp.MustCompile(`.*\/(\d+)`)
	var revote *regexp.Regexp

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
			"Variable": "SF_CONNECTOR=" + oID, // variable to track originated call
			"CallerID": ext,
			"Async":    "true",
		})
		if err != nil {
			cLog.Error(err)
		} else if res["Response"] != "Success" {
			cLog.Error(res["Message"])
		} else {
			cLog.Debug("New originate request")
		}
	}

	var connector connect.Connecter

	if cfg.B24.URL != "" {
		log.Info("Using Bitrix24 Connector")

		var sForm []connect.SForm
		if cfg.B24.hasSForm {
			for _, v := range cfg.B24.SForm {
				sForm = append(sForm, connect.SForm{R: v.R, Repl: v.Repl})
			}
		}

		connector = connect.NewB24Connector(cfg.B24.URL, cfg.B24.Token, cfg.B24.Addr, originate, sForm)
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

			cID, ok := formatNum(e["ConnectedLineNum"], true)
			if !ok {
				log.WithFields(log.Fields{"lid": e["Linkedid"], "cid": e["CallerIDNum"]}).Warn("Unknown outgoing CallerID")
				return
			}

			// register outbound call
			cdr[e["Linkedid"]] = &connect.Call{
				LID:      e["Linkedid"],
				Dir:      connect.Out,
				CID:      cID,
				DID:      e["CallerIDNum"],
				TimeCall: time.Now(),
				TimeDial: time.Now(),
				Ch:       e["Channel"],
				ChDest:   e["DestChannel"],
				Log:      log.WithField("lid", e["Linkedid"]),
			}

			// fmt.Println(e)
			cdr[e["Linkedid"]].Log.Debug("New outgoing call")

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

			if !c.O {
				go connector.Start(c)
				go connector.Dial(c, rext[1])
			}

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
			go connector.End(c)

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
		if !ok { // VarSet for originated call
			if e["Exten"] == "failed" {
				return
			}

			cdr[e["Linkedid"]] = &connect.Call{
				LID:      e["Linkedid"],
				TimeCall: time.Now(),
				Dir:      connect.Out,
				Ch:       e["Channel"],
				O:        true,
				Log:      log.WithField("lid", e["Linkedid"]),
			}

			log.WithField("oid", e["Value"]).Debug("New originated call")
			connector.OrigStart(cdr[e["Linkedid"]], e["Value"])

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
