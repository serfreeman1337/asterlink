package suitecrm

import (
	"time"

	"github.com/serfreeman1337/asterlink/connect"
)

const mysqlFormat = "2006-01-02 15:04:05"

func (s *suitecrm) Start(c *connect.Call) {
	s.ent[c.LID] = &entity{cID: c.CID, log: s.log.WithField("lid", c.LID)}
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
	if contact, _, err := s.findContact(c.CID); err == nil {
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
