package suitecrm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

const mysqlFormat = "2006-01-02 15:04:05"

type entity struct {
	ID  string
	log *log.Entry
	mux sync.Mutex
}

func (e *entity) isRegistred() bool {
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

var dirDesc = map[connect.Direction]string{connect.In: "Inbound", connect.Out: "Outbound"}

// Config struct
type Config struct {
	URL          string
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type suitecrm struct {
	cfg       *Config
	log       *log.Entry
	token     string
	tokenTime time.Time
	mux       sync.Mutex
	ent       map[string]*entity
	extUID    map[string]string
}

func (s *suitecrm) Init() {
	s.getUsers()
}

func (s *suitecrm) Start(c *connect.Call) {
	s.ent[c.LID] = &entity{log: s.log.WithField("lid", c.LID)}
	e := s.ent[c.LID]

	e.mux.Lock()

	// create a new Call record
	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "Calls",
			"attributes": map[string]interface{}{
				"name":                     c.CID,
				"direction":                dirDesc[c.Dir],
				"status":                   "Planned",
				"duration_hours":           0,
				"duration_minutes":         0,
				"asterlink_call_seconds_c": 0,
				"asterlink_cid_c":          c.CID,
				"asterlink_uid_c":          c.LID,
				"asterlink_did_c":          c.DID,
				"date_start":               time.Now().UTC().Format(mysqlFormat),
			},
		},
	}
	var r struct {
		Data struct {
			ID string
		}
	}

	err := s.rest("POST", "module", params, &r)
	// TODO: ERROR HANDLING!
	if err != nil {
		delete(s.ent, c.LID)
		e.mux.Unlock()
		return
	}

	e.ID = r.Data.ID
	e.log.WithField("id", e.ID).Debug("Call registred")
	e.mux.Unlock()

	// find call contact
	if contact, err := s.findContact(c.CID); err == nil {
		// link call record with contact
		params = map[string]interface{}{
			"data": map[string]string{
				"type": "Contacts",
				"id":   contact,
			},
		}
		err = s.rest("POST", "module/Calls/"+e.ID+"/relationships", params, nil)
	}
}

func (s *suitecrm) OrigStart(c *connect.Call, oID string) {
}

func (s *suitecrm) Dial(c *connect.Call, ext string) {
}

func (s *suitecrm) StopDial(c *connect.Call, ext string) {
}

func (s *suitecrm) Answer(c *connect.Call, ext string) {
	uID, ok := s.extUID[c.Ext]
	if !ok {
		return
	}
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	// assign answered user to call record
	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "Calls",
			"id":   e.ID,
			"attributes": map[string]string{
				"assigned_user_id": uID,
			},
		},
	}
	s.rest("PATCH", "module", params, nil)
}

func (s *suitecrm) End(c *connect.Call) {
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}
	defer delete(s.ent, c.LID)

	var (
		d      time.Duration
		status string
		uID    string
	)
	if !c.TimeAnswer.IsZero() {
		d = time.Since(c.TimeAnswer)
		status = "Held"
		uID = s.extUID[c.Ext]
	} else {
		d = time.Since(c.TimeCall)
		// status = "Not Held"
	}

	hh := int(d.Hours())
	d = d - time.Duration(hh)*time.Hour
	mm := int(d.Minutes())
	d = d - time.Duration(mm)*time.Minute
	ss := int(d.Seconds())

	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "Calls",
			"id":   e.ID,
			"attributes": map[string]interface{}{
				"status":                   status,
				"duration_hours":           hh,
				"duration_minutes":         mm,
				"asterlink_call_seconds_c": ss,
				"date_end":                 time.Now().UTC().Format(mysqlFormat),
				"assigned_user_id":         uID,
			},
		},
	}

	s.rest("PATCH", "module", params, nil)
}

func (s *suitecrm) findContact(phone string) (id string, err error) {
	//
	// search for contacts
	//
	cLog := s.log.WithField("phone", phone)
	cLog.Debug("Contact search")

	params := url.Values{}
	params.Add("fields[Contacts]", "id")
	params.Add("filter[operator]", "or")
	params.Add("filter[phone_mobile][eq]", phone)
	params.Add("filter[phone_work][eq]", phone)

	var r struct {
		Data []struct {
			ID string
		}
	}
	err = s.rest("GET", "module/Contacts?"+params.Encode(), nil, &r)
	if err != nil {
		return
	}

	if len(r.Data) == 0 {
		cLog.Debug("Not found")
		err = errors.New("Contact not found")

		return
	}

	cLog.WithField("contact", r.Data[0].ID).Debug("Found")
	id = r.Data[0].ID

	return
}

func (s *suitecrm) getUsers() (err error) {
	params := url.Values{}
	params.Add("fields[Users]", "id,asterlink_ext_c")
	// params.Add("filter[operator]", "and")
	// params.Add("filter[asterlink_ext_c][neq]", "")

	var r struct {
		Data []struct {
			ID         string
			Attributes struct {
				Ext string `json:"asterlink_ext_c"`
			}
		}
	}
	err = s.rest("GET", "module/Users?"+params.Encode(), nil, &r)
	if err != nil {
		return
	}

	for _, u := range r.Data {
		if u.Attributes.Ext != "" {
			s.extUID[u.Attributes.Ext] = u.ID
		}
	}

	s.log.WithField("users", s.extUID).Info("User list updated")

	return
}

func (s *suitecrm) refreshToken() (err error) {
	s.mux.Lock()
	s.token = ""

	type tokenRes struct {
		Token  string `json:"access_token"`
		Expire int64  `json:"expires_in"`
	}
	var r tokenRes
	err = s.rest("POST", "access_token", map[string]string{
		"grant_type":    "client_credentials", // TODO: implement password grant_type
		"client_id":     s.cfg.ClientID,
		"client_secret": s.cfg.ClientSecret,
	}, &r)

	if err != nil {
		s.mux.Unlock()
		return
	}

	s.token = r.Token
	s.tokenTime = time.Now().Add(time.Second * time.Duration(r.Expire))

	s.mux.Unlock()

	s.log.Debug("Token has been updated")
	return
}

// REST API
func (s *suitecrm) rest(method string, endpoint string, params interface{}, result interface{}) (err error) {
	isTokenReq := (endpoint == "access_token")
	var url string

	if !isTokenReq {
		s.mux.Lock()
		s.mux.Unlock()

		// request a new token if we don't have any or token is expired
		if s.token == "" || time.Now().After(s.tokenTime) {
			err = s.refreshToken()
			if err != nil {
				return
			}
		}

		url = s.cfg.URL + "V8/" + endpoint
	} else {
		url = s.cfg.URL + endpoint
	}

	var req *http.Request

	if params != nil {
		data, derr := json.Marshal(params)

		if derr != nil {
			err = derr
			s.log.Error(err)
			return
		}

		s.log.WithFields(log.Fields{"method": method, "url": url}).Trace(params)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		s.log.WithFields(log.Fields{"method": method, "url": url}).Trace()
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		s.log.Error(err)
		return
	}

	req.Header.Set("Accept", "application/vnd.api+json")

	if params != nil {
		req.Header.Set("Content-type", "application/vnd.api+json")
	}

	if !isTokenReq {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Error(err)
		return
	}

	s.log.WithFields(log.Fields{"method": method, "url": url}).Trace(res.Status)
	// expired token, request a new one and try again
	if !isTokenReq && res.StatusCode == http.StatusUnauthorized {
		res.Body.Close()
		err = s.refreshToken()
		if err != nil {
			return
		}

		req.Header.Set("Authorization", "Bearer "+s.token)

		res, err = http.DefaultClient.Do(req)
		if err != nil {
			s.log.Error(err)
			return
		}
	}

	defer res.Body.Close()

	// TODO: remove debug!
	// body, err := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		err = errors.New(res.Status)
		s.log.Error(err)
		return
	}

	if result != nil {
		// err = json.Unmarshal(body, &result)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			s.log.Error(err)
			return
		}
	}

	return
}

// NewSuiteCRMConnector func
func NewSuiteCRMConnector(cfg *Config, originate connect.OrigFunc) connect.Connecter {
	s := &suitecrm{
		cfg: cfg,
		log: log.WithField("suite", true),
		// token:     "",
		// tokenTime: time.Now().Add(1 * time.Hour),
		ent:    make(map[string]*entity),
		extUID: make(map[string]string),
	}
	s.cfg.URL += "Api/"

	return s
}
