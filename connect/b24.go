package connect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type b24 struct {
	url       string
	token     string
	addr      string
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

func (l *b24) Init(org OrigFunc) {
	l.originate = org
	l.eUID = make(map[string]string)
	l.ent = make(map[string]*entity)

	l.getUsers()

	if l.addr != "" {
		http.HandleFunc("/assigned/", l.apiAssignedHandler)
		http.HandleFunc("/originate/", l.apiOriginateHandler)
		go func() {
			l.log.WithField("addr", l.addr).Info("Enabling web server")
			err := http.ListenAndServe(l.addr, nil)
			if err != nil {
				l.log.Fatal(err)
			}
		}()
	}
}

func (l *b24) OrigStart(c *Call, oID string) {
	l.ent[c.LID] = &entity{ID: oID, log: l.log.WithField("lid", c.LID)}
}

func (l *b24) apiOriginateHandler(w http.ResponseWriter, r *http.Request) {
	cLog := l.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("methid", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseForm()

	if r.FormValue("auth[application_token]") != l.token {
		cLog.WithField("remote_addr", r.RemoteAddr).Warn("Invalid webhook token")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ext, ok := l.uIDtoExt(r.FormValue("data[USER_ID]"))
	if !ok {
		cLog.WithField("uid", r.FormValue("data[USER_ID]")).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	l.originate(ext, r.FormValue("data[PHONE_NUMBER_INTERNATIONAL]"), r.FormValue("data[CALL_ID]"))

	w.WriteHeader(http.StatusBadRequest)
}

func (l *b24) apiAssignedHandler(w http.ResponseWriter, r *http.Request) {
	cLog := l.log.WithField("api", "assigned")
	req := strings.Split(r.RequestURI, "/")[1:]

	if len(req[1]) == 0 {
		cLog.WithField("path", r.RequestURI).Warn("Incorrect RequestURI")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, ok := l.ent[req[1]]
	if !ok || !isEntRegistred(e) || e.aID == "" {
		cLog.WithField("lid", req[1]).Warn("Call not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext, ok := l.uIDtoExt(e.aID)
	if !ok {
		cLog.WithField("uid", e.aID).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (l *b24) uIDtoExt(uID string) (string, bool) {
	for k, v := range l.eUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}

func (l *b24) Start(c *Call) {
	l.ent[c.LID] = &entity{log: l.log.WithField("lid", c.LID)}
	e := l.ent[c.LID]

	e.mux.Lock()

	params := map[string]string{
		"USER_ID":      "1",
		"PHONE_NUMBER": formatCID(c.CID),
		"TYPE":         getCallType(c.Dir),
		"LINE_NUMBER":  c.DID,
		"CRM_CREATE":   "1",
	}

	if contact, err := l.findContact(params["PHONE_NUMBER"]); err == nil {
		for k, v := range contact {
			if k == "ASSIGNED_BY_ID" {
				e.aID = v
			}

			params[k] = v
		}

		delete(params, "CRM_CREATE")
	}

	res, err := l.req("telephony.externalcall.register", params)
	// TODO: ERROR HANDLING!!!
	if err != nil {
		delete(l.ent, c.LID)
		e.mux.Unlock()

		return
	}
	r := toRes(res)

	e.ID = r["CALL_ID"].(string)
	e.log.WithField("id", e.ID).Debug("Call registred")

	e.mux.Unlock()
}

func (l *b24) Dial(c *Call, ext string) {
	handleDial(l, c, ext, true)
}

func (l *b24) StopDial(c *Call, ext string) {
	handleDial(l, c, ext, false)
}

func (l *b24) Answer(c *Call, ext string) {
}

func (l *b24) End(c *Call) {
	e, ok := l.ent[c.LID]
	if !ok || !isEntRegistred(e) {
		return
	}
	defer delete(l.ent, c.CID)

	uID, ok := l.eUID[c.Ext]
	if !ok {
		// TODO: default user ID
		uID = "1"
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

	res, err := l.req("telephony.externalcall.finish", params)

	// TODO: HANDLE ERROR!!!!
	if err != nil {
		return
	}

	r := toRes(res)

	if r["CALL_FAILED_CODE"] == "304" && r["CRM_ENTITY_TYPE"] == "LEAD" {
		l.req("crm.lead.update", map[string]interface{}{
			"id": fmt.Sprintf("%.0f", r["CRM_ENTITY_ID"].(float64)),
			"fields": map[string]string{
				"TITLE":     r["PHONE_NUMBER"].(string) + " - Пропущенный звонок",
				"STATUS_ID": "NEW",
			},
			"params": map[string]string{"REGISTER_SONET_EVENT": "Y"},
		})
	}
	// delete(ent, c.LID)
}

func handleDial(l *b24, c *Call, ext string, isDial bool) {
	e, ok := l.ent[c.LID]
	if !ok || !isEntRegistred(e) {
		return
	}

	uID, ok := l.eUID[ext]
	if !ok {
		e.log.WithField("ext", ext).Warn("Cannot find user id for extension")
		return
	}

	method := "telephony.externalcall."

	if isDial {
		method += "show"
	} else {
		method += "hide"
	}

	l.req(method, map[string]string{
		"CALL_ID": e.ID,
		"USER_ID": uID,
	})
}

func (l *b24) getUsers() {
	l.log.Info("Requesting new user list")
	res, err := l.req("user.get", map[string]map[string]string{
		"filter": {"USER_TYPE": "employee"},
	})
	if err != nil {
		l.log.Error(err)
		return
	}

	for _, v := range toList(res) {
		if v["UF_PHONE_INNER"] != nil {
			l.eUID[v["UF_PHONE_INNER"].(string)] = v["ID"].(string)
		}
	}

	l.log.WithField("users", l.eUID).Info("User list updated")
}

func (l *b24) req(method string, params interface{}) (result interface{}, err error) {
	bytRep, err := json.Marshal(params)
	if err != nil {
		l.log.Error(err)
		return
	}
	l.log.WithFields(log.Fields{"method": method}).Trace(params)
	resp, err := http.Post(l.url+method+"/", "application/json", bytes.NewBuffer(bytRep))

	if err != nil {
		l.log.Error(err)
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		l.log.Error(err)
		return
	}

	var v map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v)
	result = v["result"]

	l.log.WithFields(log.Fields{"method": method}).Trace(result)

	return
}

func (l *b24) findContact(phone string) (map[string]string, error) {
	pForm := []string{phone, "+" + phone}

	for _, p := range pForm {
		cLog := l.log.WithField("phone", p)
		cLog.Debug("Contact search")
		res, err := l.req("telephony.externalCall.searchCrmEntities", map[string]string{
			"PHONE_NUMBER": p,
		})

		r := toList(res)

		if err != nil || len(r) == 0 {
			cLog.Debug("Not found")
			continue
		}

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
func NewB24Connector(url string, token string, addr string) Connecter {
	l := &b24{url: url, token: token, addr: addr, log: log.WithField("b24", true)}
	return l
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

func formatCID(phone string) string {
	return phone
}

func getCallType(d Direction) string {
	if d == Out {
		return "1"
	}

	return "2"
}
