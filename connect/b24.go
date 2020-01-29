package connect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// B24Config struct
type B24Config struct {
	IsInvalidSSL bool   `yaml:"ignore_invalid_ssl"`
	Addr         string `yaml:"webhook_endpoint_addr"`
	URL          string `yaml:"webhook_url"`
	Token        string `yaml:"webhook_originate_token"`
	RecUp        string `yaml:"rec_upload"`
	HasFindForm  bool
	FindForm     []struct {
		R    *regexp.Regexp
		Expr string
		Repl string
	} `yaml:"search_format"`
	DefUID int `yaml:"default_user"`
}

type b24 struct {
	cfg       *B24Config
	eUID      map[string]int
	ent       map[string]*b24Entity
	originate OrigFunc
	log       *log.Entry
	netClient *http.Client
}

type b24ContactInfo struct {
	Type     string `json:"CRM_ENTITY_TYPE"`
	ID       int    `json:"CRM_ENTITY_ID"`
	Assigned int    `json:"ASSIGNED_BY_ID"`
}

type b24Entity struct {
	ID  string
	cID string
	log *log.Entry
	mux sync.Mutex
}

func (e *b24Entity) isRegistred() bool {
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

func (b *b24) Init() {
	b.updateUsers()

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
	b.ent[c.LID] = &b24Entity{ID: oID, cID: c.CID, log: b.log.WithField("lid", c.LID)}
}

func (b *b24) Start(c *Call) {
	b.ent[c.LID] = &b24Entity{cID: c.CID, log: b.log.WithField("lid", c.LID)}
	e := b.ent[c.LID]

	e.mux.Lock()

	uID, ok := b.eUID[c.Ext]
	if !ok {
		uID = b.cfg.DefUID
	}

	var params struct {
		UID   int    `json:"USER_ID"`
		Phone string `json:"PHONE_NUMBER"`
		Type  int    `json:"TYPE"`
		DID   string `json:"LINE_NUMBER"`
	}

	params.UID = uID
	params.Phone = c.CID
	params.DID = c.DID

	if c.Dir == Out {
		params.Type = 1
	} else {
		params.Type = 2
	}

	var r struct {
		Result struct {
			ID string `json:"CALL_ID"`
		}
	}
	err := b.req("telephony.externalcall.register", params, &r)
	// TODO: ERROR HANDLING!!!
	if err != nil {
		delete(b.ent, c.LID)
		e.mux.Unlock()

		return
	}

	e.ID = r.Result.ID
	e.log.WithField("id", e.ID).Debug("Call registred")

	e.mux.Unlock()
}

func (b *b24) Dial(c *Call, ext string) {
	b.handleDial(c, ext, true)
}

func (b *b24) StopDial(c *Call, ext string) {
	b.handleDial(c, ext, false)
}

func (b *b24) Answer(c *Call, ext string) {
}

func (b *b24) End(c *Call) {
	e, ok := b.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}
	defer delete(b.ent, c.LID)

	uID, ok := b.eUID[c.Ext]
	if !ok {
		uID = b.cfg.DefUID
	}

	var params struct {
		ID     string `json:"CALL_ID"`
		UID    int    `json:"USER_ID"`
		Dur    int    `json:"DURATION"`
		Status string `json:"STATUS_CODE"`
		Vote   int    `json:"VOTE,omitempty"`
	}

	params.ID = e.ID
	params.UID = uID

	if !c.TimeAnswer.IsZero() {
		params.Dur = int(time.Since(c.TimeAnswer).Seconds())
		params.Status = "200"
	} else {
		params.Dur = int(time.Since(c.TimeCall).Seconds())
		params.Status = "304"
	}

	if c.Vote != "" && c.Vote != "-" {
		params.Vote, _ = strconv.Atoi(c.Vote)
	}

	err := b.req("telephony.externalcall.finish", params, nil)
	// TODO: HANDLE ERROR!!!!
	if err != nil {
		return
	}

	// upload recording
	if b.cfg.RecUp != "" && !c.TimeAnswer.IsZero() && c.Rec != "" {
		file := path.Base(c.Rec)
		url := b.cfg.RecUp + c.Rec

		e.log.WithFields(log.Fields{url: url}).Debug("Attaching call record")
		b.req("telephony.externalCall.attachRecord", map[string]string{
			"CALL_ID":    e.ID,
			"FILENAME":   file,
			"RECORD_URL": url,
		}, nil)
	}
}

func (b *b24) handleDial(c *Call, ext string, isDial bool) {
	e, ok := b.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	uID, ok := b.eUID[ext]
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

	var params struct {
		ID  string `json:"CALL_ID"`
		UID int    `json:"USER_ID"`
	}
	params.ID = e.ID
	params.UID = uID

	b.req(method, params, nil)
}

func (b *b24) updateUsers() {
	// TODO: Check how it works with large user list!
	var r struct {
		Result []struct {
			ID    int    `json:"ID,string"`
			Phone string `json:"UF_PHONE_INNER"`
		}
	}
	err := b.req("user.get", map[string]map[string]string{
		"filter": {"USER_TYPE": "employee"},
	}, &r)

	if err != nil {
		b.log.Error("Failed to update users list")
		return
	}

	for _, v := range r.Result {
		if v.Phone != "" {
			b.eUID[v.Phone] = v.ID
		}
	}

	b.log.WithField("users", b.eUID).Info("User list updated")
}

func (b *b24) apiOriginateHandler(w http.ResponseWriter, r *http.Request) {
	cLog := b.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("method", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseForm()

	if r.FormValue("auth[application_token]") != b.cfg.Token {
		cLog.WithField("remote_addr", r.RemoteAddr).Warn("Invalid webhook token")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	uID, _ := strconv.Atoi(r.FormValue("data[USER_ID]"))
	ext, ok := b.uIDtoExt(uID)
	if !ok {
		cLog.WithField("uid", uID).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b.originate(ext, r.FormValue("data[PHONE_NUMBER_INTERNATIONAL]"), r.FormValue("data[CALL_ID]"))
	w.WriteHeader(http.StatusOK)
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
	if !ok || !e.isRegistred() {
		cLog.WithField("lid", req[1]).Warn("Call not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// search for contact
	contact, err := b.findContact(e.cID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext, ok := b.uIDtoExt(contact.Assigned)
	if !ok {
		cLog.WithField("uid", contact.Assigned).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (b *b24) findContact(phone string) (*b24ContactInfo, error) {
	var num []string

	if !b.cfg.HasFindForm {
		num = []string{phone}
	} else {
		for _, ff := range b.cfg.FindForm {
			if !ff.R.MatchString(phone) {
				continue
			}

			num = append(num, ff.R.ReplaceAllString(phone, ff.Repl))
		}
	}

	for _, p := range num {
		cLog := b.log.WithField("phone", p)
		cLog.Debug("Contact search")

		var r struct {
			Result []b24ContactInfo
		}
		err := b.req("telephony.externalCall.searchCrmEntities", map[string]string{
			"PHONE_NUMBER": p,
		}, &r)
		if err != nil {
			continue
		}

		if len(r.Result) == 0 {
			cLog.Debug("Not found")
			continue
		}

		// TODO: handle multiple crm entitites
		cLog.WithField("contact", r.Result[0]).Debug("Found")

		return &r.Result[0], nil
	}
	return nil, errors.New("Contact not found")
}

func (b *b24) uIDtoExt(uID int) (string, bool) {
	for k, v := range b.eUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}

func (b *b24) req(method string, params interface{}, result interface{}) (err error) {
	data, err := json.Marshal(params)
	if err != nil {
		b.log.Error(err)
		return
	}
	b.log.WithFields(log.Fields{"method": method}).Trace(params)

	res, err := b.netClient.Post(b.cfg.URL+method+"/", "application/json", bytes.NewBuffer(data))
	if err != nil {
		b.log.Error(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = errors.New(res.Status)
		b.log.Error(err)
		return
	}

	if result != nil {
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			b.log.Error(err)
			return
		}

		b.log.WithFields(log.Fields{"method": method}).Trace(result)
	}

	return
}

// NewB24Connector connector
func NewB24Connector(cfg *B24Config, originate OrigFunc) Connecter {
	client := http.Client{
		Timeout: time.Second * 30,
	}

	if cfg.IsInvalidSSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
		log.Info("Ignore invalid self-signed certificates")
	}

	b := &b24{
		cfg:       cfg,
		log:       log.WithField("b24", true),
		originate: originate,
		eUID:      make(map[string]int),
		ent:       make(map[string]*b24Entity),
		netClient: &client,
	}

	return b
}
