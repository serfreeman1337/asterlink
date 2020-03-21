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
	defer e.mux.Unlock()

	var err error

	e.ID, err = s.createCallRecord(c)
	if err != nil {
		delete(s.ent, c.LID)
		return
	}

	e.log.WithField("id", e.ID).Debug("Call registred")
}

func (s *suitecrm) OrigStart(c *connect.Call, oID string) {
	s.ent[c.LID] = &entity{ID: oID, log: s.log.WithField("lid", c.LID)}
}

func (s *suitecrm) Dial(c *connect.Call, ext string) {
	if c.O { // update uid for originated call
		e, ok := s.ent[c.LID]
		if !ok || !e.isRegistred() {
			return
		}

		params := map[string]interface{}{
			"data": map[string]interface{}{
				"type": "Calls",
				"id":   e.ID,
				"attributes": map[string]string{
					"asterlink_uid_c": c.LID,
				},
			},
		}
		s.rest("PATCH", "module", params, nil)
	}
}

func (s *suitecrm) StopDial(c *connect.Call, ext string) {
}

func (s *suitecrm) Answer(c *connect.Call, ext string) {
	// update user id for incoming call on answer
	// or set DID for originated call
	if c.Dir != connect.In && !c.O {
		return
	}

	uID, ok := s.extUID[c.Ext]
	if !ok {
		return
	}
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	attr := make(map[string]string)

	if !c.O {
		attr["assigned_user_id"] = uID
	} else { // user already assigned on origate request
		attr["asterlink_did_c"] = c.DID
	}

	// assign answered user to call record
	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type":       "Calls",
			"id":         e.ID,
			"attributes": attr,
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
		// uID    string
	)
	if !c.TimeAnswer.IsZero() {
		d = time.Since(c.TimeAnswer)
		status = "Held"
		// uID = s.extUID[c.Ext]
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
				// "assigned_user_id":         uID,
			},
		},
	}

	s.rest("PATCH", "module", params, nil)
}
