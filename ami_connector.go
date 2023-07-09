package main

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/ivahaev/amigo"
	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

type numForm struct {
	R    *regexp.Regexp
	Expr string
	Repl string
}

type AmiConfig struct {
	Ami struct {
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
}

type amiConnector struct {
	ami       *amigo.Amigo
	cfg       *AmiConfig
	cdr       map[string]*connect.Call
	rec       map[string]string
	revote    *regexp.Regexp
	rechan    *regexp.Regexp
	connector connect.Connecter
}

func (a *amiConnector) Init() {
	a.ami.Connect()
}

func (a *amiConnector) Originate(ext string, dest string, oID string) {
	cLog := log.WithFields(log.Fields{"ext": ext, "dest": dest, "oID": oID})
	num, ok := a.formatNum(dest, false)

	if !ok {
		cLog.WithField("num", num).Error("Invalid number to originate call")
		return
	}

	res, err := a.ami.Action(map[string]string{
		"Action":   "Originate",
		"Channel":  "Local/" + ext + "@" + a.cfg.DP.Dial,
		"Exten":    num,
		"Context":  a.cfg.DP.Dial,
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

func (a *amiConnector) formatNum(cID string, isCid bool) (string, bool) {
	var hasForm bool
	var form []numForm
	if isCid {
		hasForm = a.cfg.Form.hasCid
		form = a.cfg.Form.Cid
	} else {
		hasForm = a.cfg.Form.hasDial
		form = a.cfg.Form.Dial
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

func (a *amiConnector) useRec(c *connect.Call, uID string, lID string) {
	// Check recording with uniqueid (when recoging is turned on extension).
	rec, ok := a.rec[uID]

	if !ok { // Then look for recording with linkedid (when recording is turned on queue etc.).
		rec, ok = a.rec[lID]
		uID = lID
	}

	if !ok {
		return
	}

	c.Rec = rec
	delete(a.rec, uID)

	c.Log.WithField("rec", c.Rec).Debug("Set recording file")
}

func (a *amiConnector) isContext(context string, of []string) bool {
	for _, v := range of {
		if context == v {
			return true
		}
	}

	return false
}

func (a *amiConnector) onConnect(message string) {
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
	filter := "(Event: Newchannel.*[[:cntrl:]]Context: " + fexpr(a.cfg.DP.In) + "[[:cntrl:]]"
	filter += "|Event: (DialBegin|DialEnd).*[[:cntrl:]]Context: " + fexpr(append(a.cfg.DP.Ext, a.cfg.DP.Out...)) + "[[:cntrl:]]"
	filter += "|Event: Newexten.*[[:cntrl:]]"

	if a.cfg.DP.Vote == "" {
		filter += "Application: MixMonitor"
	} else {
		filter += "(Application: MixMonitor|Context: " + a.cfg.DP.Vote + "[[:cntrl:]].*AppData: CDR\\(userfield\\)=\".*rate.*\")"
		filter += "|Variable: IVR_CONTEXT.*Value: " + a.cfg.DP.Vote
		a.revote = regexp.MustCompile(`CDR\(userfield\)="rate:(\d+)"`)
	}

	filter += "|Event: (Hangup|BlindTransfer)|Variable: SF_CONNECTOR)"

	// let the Asterisk filter events for us
	res, err := a.ami.Action(map[string]string{"Action": "Filter", "Operation": "Add", "Filter": filter})
	if err != nil {
		log.WithField("filter", filter).Fatal(err)
	} else if res["Response"] != "Success" {
		log.WithField("filter", filter).Fatal(res["Message"])
	}
}

func (a *amiConnector) onError(message string) {
	log.Error(message)
}

// incoming call
func (a *amiConnector) onNewchannel(e map[string]string) {
	if /*_, ok := IncomingContext[e["Context"]]; !ok || */ e["Exten"] == "s" { // TODO: review this 's'!
		return
	}

	if _, ok := a.cdr[e["Linkedid"]]; ok {
		log.WithField("lid", e["Linkedid"]).Warn("Already tracked")

		return
	}

	// DID Limit
	if a.cfg.Form.hasDid {
		if _, ok := a.cfg.Form.DID[e["Exten"]]; !ok {
			log.WithFields(log.Fields{"lid": e["Linkedid"], "did": e["Exten"]}).Debug("Skip incoming call for DID")
			return
		}
	}

	// skip anonymous calls ?
	if e["CallerIDNum"] == "<unknown>" {
		log.WithFields(log.Fields{"lid": e["Linkedid"]}).Warn("Anonymous CallerID")
		return
	}

	cID, ok := a.formatNum(e["CallerIDNum"], true)
	if !ok {
		log.WithFields(log.Fields{"lid": e["Linkedid"], "cid": e["CallerIDNum"]}).Warn("Unknown incoming CallerID")
		return
	}

	// register incoming call
	a.cdr[e["Linkedid"]] = &connect.Call{
		LID:      e["Linkedid"],
		Dir:      connect.In,
		CID:      cID,
		DID:      e["Exten"],
		TimeCall: time.Now(),
		Ch:       e["Channel"],
		Log:      log.WithField("lid", e["Linkedid"]),
	}
	a.cdr[e["Linkedid"]].Log.Debug("New incoming call")

	go a.connector.Start(a.cdr[e["Linkedid"]])
}

// operator dial
func (a *amiConnector) onDialBegin(e map[string]string) {
	c, ok := a.cdr[e["Linkedid"]]

	if !ok {
		// TODO: review
		if !a.isContext(e["Context"], a.cfg.DP.Out) {
			return
		}

		cID, ok := a.formatNum(e["DestCallerIDNum"], true)
		if !ok {
			log.WithFields(log.Fields{"lid": e["Linkedid"], "cid": e["DestCallerIDNum"]}).Warn("Unknown outgoing CallerID")
			return
		}

		rext := a.rechan.FindStringSubmatch(e["Channel"])
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
		a.cdr[e["Linkedid"]] = c

		c.Log.Debug("New outgoing call")

		go func() {
			a.connector.Start(c)
			a.connector.Dial(c, rext[1])
		}()

		return
	} else if c.O { // Originated call
		if a.isContext(e["Context"], a.cfg.DP.Ext) {
			c.Ch = e["DestChannel"]
			c.Ext = e["DestCallerIDNum"]

			c.Log.WithField("ext", c.Ext).Debug("From")
			a.useRec(c, e["Uniqueid"], e["Linkedid"])

			go a.connector.Dial(c, c.Ext)
		} else if a.isContext(e["Context"], a.cfg.DP.Out) {
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
	go a.connector.Dial(c, e["DestCallerIDNum"])
}

// operator answer
func (a *amiConnector) onDialEnd(e map[string]string) {
	c, ok := a.cdr[e["Linkedid"]]

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

		a.useRec(c, e["Uniqueid"], e["Linkedid"])

		c.Log.WithField("ext", c.Ext).Debug("Answer")
		go a.connector.Answer(c, c.Ext)
		break
	case connect.Out:
		if c.O && !a.isContext(e["Context"], a.cfg.DP.Out) { // Originated call
			return
		}

		// TODO: the heck am i doing ?
		if a.isContext(e["Context"], a.cfg.DP.Ext) && c.ChDest == e["Channel"] {
			// BlindTransfer hack ?
			rext := a.rechan.FindStringSubmatch(e["DestChannel"])
			if len(rext) == 0 {
				return
			}

			c.Ext = rext[1]
			c.Log.WithField("ext", c.Ext).Debug("Answer 2")
			a.useRec(c, e["Uniqueid"], e["Linkedid"])
			go a.connector.Answer(c, c.Ext)

			return
		}

		rext := a.rechan.FindStringSubmatch(e["Channel"])
		if len(rext) == 0 {
			return
		}

		c.TimeAnswer = time.Now()
		c.Ext = rext[1]

		c.Log.WithField("ext", c.Ext).Debug("Dial 2")
		a.useRec(c, e["Uniqueid"], e["Linkedid"])

		go a.connector.Answer(c, rext[1])

		break
	}
}

// call recording and vote
func (a *amiConnector) onNewexten(e map[string]string) {
	if e["Application"] == "MixMonitor" { // recording
		a.rec[e["Uniqueid"]] = strings.Split(e["AppData"], ",")[0]
		log.WithFields(log.Fields{"lid": e["Linkedid"], "uid": e["Uniqueid"], "rec": a.rec[e["Uniqueid"]]}).Debug("MixMonitor")
	} else { // vote
		c, ok := a.cdr[e["Linkedid"]]
		if !ok {
			return
		}
		appData := e["AppData"][:len(e["AppData"])-1] // 3 IQ WHITESPACE TRIM
		c.Vote = a.revote.ReplaceAllString(appData, "$1")
		c.Log.WithField("vote", c.Vote).Debug("Call voted")
	}
}

func (a *amiConnector) onBlindTransfer(e map[string]string) {
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

	c, ok := a.cdr[lID]
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
	go a.connector.StopDial(c, c.Ext)
}

// call finish
func (a *amiConnector) onHangup(e map[string]string) {
	_, ok := a.rec[e["Uniqueid"]]

	if ok {
		log.WithFields(log.Fields{"lid": e["Linkedid"], "uid": e["Uniqueid"]}).Debug("Clear recording")
		delete(a.rec, e["Uniqueid"])
	}

	c, ok := a.cdr[e["Linkedid"]]

	if !ok {
		return
	}

	if e["Channel"] == c.Ch || e["Channel"] == c.ChDest {
		// call marked for vote, keep tracking after operator hangup
		if c.Vote == "-" && !c.TimeAnswer.IsZero() && e["Channel"] == c.ChDest {
			go a.connector.StopDial(c, c.Ext)

			c.ChDest = ""
			c.Log.Debug("Wait for vote")

			return
		}

		c.Log.Debug("Call finished")
		go a.connector.End(c, e["Cause"])

		delete(a.cdr, e["Linkedid"])
	} else {
		if e["Context"] == a.cfg.DP.Dial && e["ChannelState"] == "5" {
			c.Log.WithField("ext", e["CallerIDNum"]).Debug("Dial stop")
			go a.connector.StopDial(c, e["CallerIDNum"])
		}
	}
}

// originated call
func (a *amiConnector) onVarSet(e map[string]string) {
	c, ok := a.cdr[e["Linkedid"]]
	if !ok {
		// VarSet for originated call
		if e["Variable"] != "SF_CONNECTOR" || e["Exten"] == "failed" {
			return
		}

		// split tracker variable (see Originate request)
		r := strings.Split(e["Value"], "|")

		a.cdr[e["Linkedid"]] = &connect.Call{
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
		a.connector.OrigStart(a.cdr[e["Linkedid"]], r[0])

		return
	}

	if a.cfg.DP.Vote == "" || e["Value"] != a.cfg.DP.Vote {
		return
	}

	// Mark call for vote
	c.Vote = "-"
}

func (a *amiConnector) SetConnector(cc connect.Connecter) {
	a.connector = cc
}

func NewAmiConnector(cfg *AmiConfig) (a *amiConnector, err error) {
	if cfg.Ami.Host == "" || cfg.Ami.User == "" || cfg.Ami.Pass == "" {
		err = errors.New("AMI settings are missing from config file")
		return
	} else if len(cfg.DP.In) == 0 || len(cfg.DP.Out) == 0 || len(cfg.DP.Ext) == 0 || cfg.DP.Dial == "" {
		err = errors.New("DialPlan configuration are missing from config file")
		return
	}

	if cfg.Ami.Port == "" {
		cfg.Ami.Port = "5038"
	}

	if len(cfg.Form.Cid) != 0 {
		for i, v := range cfg.Form.Cid {
			if cfg.Form.Cid[i].R, err = regexp.Compile(v.Expr); err != nil {
				return
			}
		}
		cfg.Form.hasCid = true
	}

	if len(cfg.Form.Dial) != 0 {
		for i, v := range cfg.Form.Dial {
			if cfg.Form.Dial[i].R, err = regexp.Compile(v.Expr); err != nil {
				return
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

	settings := &amigo.Settings{Host: cfg.Ami.Host, Username: cfg.Ami.User, Password: cfg.Ami.Pass, Port: cfg.Ami.Port}

	a = &amiConnector{
		cfg:    cfg,
		ami:    amigo.New(settings),
		cdr:    make(map[string]*connect.Call),
		rec:    make(map[string]string),
		rechan: regexp.MustCompile(`.*\/(\d+)`),
	}

	a.ami.On("connect", a.onConnect)
	a.ami.On("error", a.onError)
	a.ami.RegisterHandler("Newchannel", a.onNewchannel)
	a.ami.RegisterHandler("DialBegin", a.onDialBegin)
	a.ami.RegisterHandler("DialEnd", a.onDialEnd)
	a.ami.RegisterHandler("Newexten", a.onNewexten)
	a.ami.RegisterHandler("BlindTransfer", a.onBlindTransfer)
	a.ami.RegisterHandler("Hangup", a.onHangup)
	a.ami.RegisterHandler("VarSet", a.onVarSet)

	return
}
