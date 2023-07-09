package suitecrm

import (
	"time"

	"github.com/serfreeman1337/asterlink/connect"
)

const mysqlFormat = "2006-01-02 15:04:05"

func (s *suitecrm) Start(c *connect.Call) {
	s.ent[c.LID] = &entity{
		Dir: c.Dir,
		DID: c.DID,
		CID: c.CID,

		log: s.log.WithField("lid", c.LID),
	}

	e := s.ent[c.LID]

	e.mux.Lock()
	defer e.mux.Unlock()

	if err := s.createCallRecord(c, e); err != nil {
		delete(s.ent, c.LID)
		return
	}

	e.log.WithField("id", e.ID).Debug("Call registred")
}

func (s *suitecrm) OrigStart(c *connect.Call, oID string) {
	e := &entity{
		ID:  oID,
		Dir: c.Dir,
		CID: c.CID,

		log: s.log.WithField("lid", c.LID),
	}
	s.ent[c.LID] = e

	// TODO: rewrite
	e.mux.Lock()

	params := map[string]interface{}{
		"id": oID,
	}
	s.rest("POST", "get_relations", params, &e.Relationships)

	e.mux.Unlock()
}

func (s *suitecrm) Dial(c *connect.Call, ext string) {
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	e.exts.Store(ext, true)
	e.TimeStamp = c.TimeDial
	s.wsBroadcast(ext, true, e)

	if c.O { // update uid for originated call
		params := map[string]interface{}{
			"id": e.ID,
			"data": map[string]interface{}{
				"asterlink_uid_c": c.LID,
			},
		}

		e.mux.Lock()
		s.rest("POST", "update_call_record", params, nil)
		e.mux.Unlock()
	}
}

func (s *suitecrm) StopDial(c *connect.Call, ext string) {
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	e.exts.Delete(ext)
	s.wsBroadcast(ext, false, e)
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

	e.TimeStamp = c.TimeAnswer
	e.IsAnswered = true
	e.DID = c.DID
	s.wsBroadcast(ext, true, e)

	// update user id for incoming call on answer
	// or set DID for originated call
	if c.Dir != connect.In && !c.O {
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
		"id":   e.ID,
		"data": attr,
	}

	e.mux.Lock()
	s.rest("POST", "update_call_record", params, nil)
	e.mux.Unlock()
}

func (s *suitecrm) End(c *connect.Call, cause string) {
	e, ok := s.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}
	defer delete(s.ent, c.LID)

	e.exts.Range(func(key interface{}, value interface{}) bool {
		e.exts.Delete(key)

		s.wsBroadcast(key.(string), false, e)

		return true
	})

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
		"id": e.ID,
		"data": map[string]interface{}{
			"status":                   status,
			"duration_hours":           hh,
			"duration_minutes":         mm,
			"asterlink_call_seconds_c": ss,
			"date_end":                 time.Now().UTC().Format(mysqlFormat),
			// "assigned_user_id":         uID,
		},
	}

	e.mux.Lock()
	s.rest("POST", "update_call_record", params, nil)
	e.mux.Unlock()
}
