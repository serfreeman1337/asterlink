package suitecrm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/serfreeman1337/asterlink/connect"
)

// type contact struct {
// 	ID         string `json:"id,omitempty"`
// 	Name       string `json:"name,omitempty"`
// 	AssignedID string `json:"assigned,omitempty"`
// }

type relation struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	AssignedID string `json:"-"`
}

func (s *suitecrm) createCallRecord(c *connect.Call, e *entity) (err error) {
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

	err = s.rest("POST", "create_call_record", attr, &e)

	// TODO: ERROR HANDLING!
	if err != nil {
		return
	}

	return
}

func (s *suitecrm) getExtUsers() (err error) {
	clear(s.extUID)
	err = s.rest("GET", "get_ext_users", nil, &s.extUID)
	if err != nil {
		return
	}

	s.log.WithField("users", s.extUID).Info("User list updated")

	return
}

// REST API
func (s *suitecrm) rest(method string, action string, params interface{}, result interface{}) (err error) {
	url := s.cfg.URL + "&action=" + action

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

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Asterlink-Token", s.cfg.EndpointToken)

	if params != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Error(err)
		return
	}

	s.log.WithFields(log.Fields{"method": method, "url": url}).Trace(res.Status)
	defer res.Body.Close()

	// TODO: remove debug!
	// body, err := ioutil.ReadAll(res.Body)
	// fmt.Println(string(body))

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
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
