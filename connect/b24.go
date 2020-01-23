package connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// B24Config struct
type B24Config struct {
	Addr        string `yaml:"webhook_endpoint_addr"`
	URL         string `yaml:"webhook_url"`
	Token       string `yaml:"webhook_originate_token"`
	RecUp       string `yaml:"rec_upload"`
	HasFindForm bool
	FindForm    []struct {
		R    *regexp.Regexp
		Expr string
		Repl string
	} `yaml:"search_format"`
	DefUID string `yaml:"default_user"`
}

type b24 struct {
	cfg       *B24Config
	eUID      map[string]string
	ent       map[string]*entity
	originate OrigFunc
	log       *log.Entry
}

type entity struct {
	ID  string
	aID string
	log *log.Entry
	mux sync.Mutex
}

func (b *b24) Init() {
	b.getUsers()

	if b.cfg.Addr != "" {
		http.HandleFunc("/assigned/", b.apiAssignedHandler)
		http.HandleFunc("/originate/", b.apiOriginateHandler)
		go func() {
			b.log.WithField("addr", b.cfg.Addr).Info("Enabling web server")
			err := http.ListenAndServe(b.cfg.Addr, nil)
			if err != nil {
				b.log.Fatal(err)
			}
		}()
	}
}

func (b *b24) OrigStart(c *Call, oID string) {
	b.ent[c.LID] = &entity{ID: oID, log: b.log.WithField("lid", c.LID)}
}

func (b *b24) apiOriginateHandler(w http.ResponseWriter, r *http.Request) {
	cLog := b.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("methid", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseForm()

	if r.FormValue("auth[application_token]") != b.cfg.Token {
		cLog.WithField("remote_addr", r.RemoteAddr).Warn("Invalid webhook token")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ext, ok := b.uIDtoExt(r.FormValue("data[USER_ID]"))
	if !ok {
		cLog.WithField("uid", r.FormValue("data[USER_ID]")).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b.originate(ext, r.FormValue("data[PHONE_NUMBER_INTERNATIONAL]"), r.FormValue("data[CALL_ID]"))

	w.WriteHeader(http.StatusBadRequest)
}

func (b *b24) apiAssignedHandler(w http.ResponseWriter, r *http.Request) {
	cLog := b.log.WithField("api", "assigned")
	req := strings.Split(r.RequestURI, "/")[1:]

	if len(req[1]) == 0 {
		cLog.WithField("path", r.RequestURI).Warn("Incorrect RequestURI")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, ok := b.ent[req[1]]
	if !ok || !isEntRegistred(e) {
		cLog.WithField("lid", req[1]).Warn("Call not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if e.aID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext, ok := b.uIDtoExt(e.aID)
	if !ok {
		cLog.WithField("uid", e.aID).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (b *b24) uIDtoExt(uID string) (string, bool) {
	for k, v := range b.eUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}

func (b *b24) Start(c *Call) {
	b.ent[c.LID] = &entity{log: b.log.WithField("lid", c.LID)}
	e := b.ent[c.LID]

	e.mux.Lock()

	uID, ok := b.eUID[c.Ext]
	if !ok {
		uID = b.cfg.DefUID
	}

	params := map[string]string{
		"USER_ID":      uID,
		"PHONE_NUMBER": c.CID,
		"TYPE":         getCallType(c.Dir),
		"LINE_NUMBER":  c.DID,
	}

	if contact, err := b.findContact(params["PHONE_NUMBER"]); err == nil {
		for k, v := range contact {
			if k == "ASSIGNED_BY_ID" {
				e.aID = v
			}

			params[k] = v
		}
	}

	res, err := b.req("telephony.externalcalb.register", params)
	// TODO: ERROR HANDLING!!!
	if err != nil {
		delete(b.ent, c.LID)
		e.mux.Unlock()

		return
	}
	r := toRes(res)

	e.ID = r["CALL_ID"].(string)
	e.log.WithField("id", e.ID).Debug("Call registred")

	e.mux.Unlock()
}

func (b *b24) Dial(c *Call, ext string) {
	handleDial(b, c, ext, true)
}

func (b *b24) StopDial(c *Call, ext string) {
	handleDial(b, c, ext, false)
}

func (b *b24) Answer(c *Call, ext string) {
}

func (b *b24) End(c *Call) {
	e, ok := b.ent[c.LID]
	if !ok || !isEntRegistred(e) {
		return
	}
	defer delete(b.ent, c.CID)

	uID, ok := b.eUID[c.Ext]
	if !ok {
		uID = b.cfg.DefUID
	}

	params := map[string]string{
		"CALL_ID": e.ID,
		"USER_ID": uID,
	}

	if !c.TimeAnswer.IsZero() {
		params["DURATION"] = fmt.Sprintf("%.0f", time.Since(c.TimeAnswer).Seconds())
		params["STATUS_CODE"] = "200"
	} else {
		params["DURATION"] = fmt.Sprintf("%.0f", time.Since(c.TimeCall).Seconds())
		params["STATUS_CODE"] = "304"
	}

	if c.Vote != "" && c.Vote != "-" {
		params["VOTE"] = c.Vote
	}

	_, err := b.req("telephony.externalcalb.finish", params)

	// TODO: HANDLE ERROR!!!!
	if err != nil {
		return
	}

	// upload recording
	if b.cfg.RecUp != "" && !c.TimeAnswer.IsZero() && c.Rec != "" {
		file := path.Base(c.Rec)
		url := b.cfg.RecUp + c.Rec

		e.log.WithFields(log.Fields{url: url}).Debug("Attaching call record")
		b.req("telephony.externalCalb.attachRecord", map[string]string{
			"CALL_ID":    e.ID,
			"FILENAME":   file,
			"RECORD_URL": url,
		})
	}

	// delete(ent, c.LID)
}

func handleDial(b *b24, c *Call, ext string, isDial bool) {
	e, ok := b.ent[c.LID]
	if !ok || !isEntRegistred(e) {
		return
	}

	uID, ok := b.eUID[ext]
	if !ok {
		e.log.WithField("ext", ext).Warn("Cannot find user id for extension")
		return
	}

	method := "telephony.externalcalb."

	if isDial {
		method += "show"
	} else {
		method += "hide"
	}

	b.req(method, map[string]string{
		"CALL_ID": e.ID,
		"USER_ID": uID,
	})
}

func (b *b24) getUsers() {
	res, err := b.req("user.get", map[string]map[string]string{
		"filter": {"USER_TYPE": "employee"},
	})
	if err != nil {
		b.log.Error(err)
		return
	}

	for _, v := range toList(res) {
		if v["UF_PHONE_INNER"] != nil {
			b.eUID[v["UF_PHONE_INNER"].(string)] = v["ID"].(string)
		}
	}

	b.log.WithField("users", b.eUID).Info("User list updated")
}

func (b *b24) req(method string, params interface{}) (result interface{}, err error) {
	bytRep, err := json.Marshal(params)
	if err != nil {
		b.log.Error(err)
		return
	}
	b.log.WithFields(log.Fields{"method": method}).Trace(params)
	resp, err := http.Post(b.cfg.URL+method+"/", "application/json", bytes.NewBuffer(bytRep))

	if err != nil {
		b.log.Error(err)
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		b.log.Error(err)
		return
	}

	var v map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v)
	result = v["result"]

	b.log.WithFields(log.Fields{"method": method}).Trace(result)

	return
}

func (b *b24) findContact(phone string) (map[string]string, error) {
	var pForm []string

	if !b.cfg.HasFindForm {
		pForm = []string{phone}
	} else {
		for _, sF := range b.cfg.FindForm {
			if !sF.R.MatchString(phone) {
				continue
			}

			pForm = append(pForm, sF.R.ReplaceAllString(phone, sF.Repl))
		}
	}

	for _, p := range pForm {
		cLog := b.log.WithField("phone", p)
		cLog.Debug("Contact search")

		res, err := b.req("telephony.externalCalb.searchCrmEntities", map[string]string{
			"PHONE_NUMBER": p,
		})
		if err != nil {
			continue
		}

		r := toList(res)

		if len(r) == 0 {
			cLog.Debug("Not found")
			continue
		}

		// TODO: handle multiple crm entitites
		cLog.WithField("contact", r[0]).Debug("Found")

		return map[string]string{
			"CRM_ENTITY_TYPE": r[0]["CRM_ENTITY_TYPE"].(string),
			"CRM_ENTITY_ID":   fmt.Sprintf("%.0f", r[0]["CRM_ENTITY_ID"].(float64)),
			"ASSIGNED_BY_ID":  fmt.Sprintf("%.0f", r[0]["ASSIGNED_BY_ID"].(float64)),
		}, nil
	}
	return nil, errors.New("Contact not found")
}

// NewB24Connector func
func NewB24Connector(cfg *B24Config, originate OrigFunc) Connecter {
	b := &b24{
		cfg:       cfg,
		log:       log.WithField("b24", true),
		originate: originate,
		eUID:      make(map[string]string),
		ent:       make(map[string]*entity),
	}

	return b
}

func toRes(res interface{}) map[string]interface{} {
	return res.(map[string]interface{})
}

func toList(res interface{}) []map[string]interface{} {
	t := res.([]interface{})
	r := make([]map[string]interface{}, len(t))

	for k, v := range t {
		r[k] = v.(map[string]interface{})
	}

	return r
}

func isEntRegistred(e *entity) bool {
	if e.ID != "" {
		return true
	}

	e.mux.Lock()
	e.mux.Unlock()

	if e.ID == "" {
		return false
	}
	return true
}

func getCallType(d Direction) string {
	if d == Out {
		return "1"
	}

	return "2"
}
