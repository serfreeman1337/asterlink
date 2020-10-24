package suitecrm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

type contact struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	AssignedID string `json:"assigned,omitempty"`
}

type relation struct {
	Module        string `json:"module"`
	ModuleName    string `json:"module_name"`
	PrimaryModule bool   `json:"-"`
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	AssignedID    string `json:"-"`
}

func (s *suitecrm) createCallRecord(c *connect.Call) (id string, relations []relation, err error) {
	attr := map[string]interface{}{
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
	}

	if c.Ext != "" {
		if uID, ok := s.extUID[c.Ext]; ok {
			attr["assigned_user_id"] = uID
		}
	}

	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "Calls",
			"attributes": attr,
		},
	}
	var r struct {
		Data struct {
			ID string
		}
	}

	err = s.rest("POST", "module", params, &r)
	// TODO: ERROR HANDLING!
	if err != nil {
		return
	}

	id = r.Data.ID

	// call relation
	for rel := range s.findRelation(c.CID) {
		relations = append(relations, rel)
		// link call record with related record
		var params map[string]interface{}
		var url string

		if rel.PrimaryModule {
			params = map[string]interface{}{
				"data": map[string]string{
					"type": "Calls",
					"id":   id,
				},
			}

			url = "module/" + rel.Module + "/" + rel.ID + "/relationships"
		} else {
			params = map[string]interface{}{
				"data": map[string]string{
					"type": rel.Module,
					"id":   rel.ID,
				},
			}

			url = "module/Calls/" + id + "/relationships"
		}

		// don't care about errors
		s.rest("POST", url, params, nil)
	}

	return
}

func (s *suitecrm) findRelation(phone string) <-chan relation {
	out := make(chan relation)

	// search for contacts
	cLog := s.log.WithField("phone", phone)
	cLog.Debug("Relation search")

	go func() {
		for _, rs := range s.cfg.Relationships {
			params := url.Values{}
			// TODO:  provide PR to suitecrm ?
			// params.Add("fields[Contacts]", "id,full_name,assigned_user_id") // full_name isn't working with fields filter

			// See issue #8366 -> https://github.com/salesagility/SuiteCRM/issues/8366
			params.Add("filter[operator]", "or")
			for _, field := range rs.PhoneFields {
				params.Add("filter["+field+"][eq]", phone)
			}

			var r struct {
				Data []struct {
					ID         string
					Attributes map[string]json.RawMessage
				}
			}
			err := s.rest("GET", "module/"+rs.Module+"?"+params.Encode(), nil, &r)
			if err != nil {
				continue
			}

			if len(r.Data) == 0 {
				continue
			}

			if _, ok := r.Data[0].Attributes[rs.NameField]; !ok {
				cLog.WithFields(log.Fields{"module": rs.Module, "name_field": rs.NameField}).Warn("Unknown name field")
				continue
			}

			assignedID := ""
			if aid, ok := r.Data[0].Attributes["assigned_user_id"]; ok {
				assignedID = strings.Trim(string(aid), "\"")
			}

			out <- relation{rs.Module, rs.ModuleName, rs.PrimaryModule,
				r.Data[0].ID,
				strings.Trim(string(r.Data[0].Attributes[rs.NameField]), "\""),
				assignedID,
			}

			if s.cfg.RelateOnce {
				break
			}
		}

		close(out)
	}()

	return out
}

func (s *suitecrm) getUsers() (err error) {
	params := url.Values{}
	params.Add("fields[Users]", "id,asterlink_ext_c")

	// TODO:  provide PR to suitecrm ?
	// you can't apply != "" filter as for 7.10.22 version
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

	// fmt.Println(s.token)
	s.log.Debug("Token has been updated")
	return
}
