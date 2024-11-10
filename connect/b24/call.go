package b24

import (
	"encoding/base64"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/serfreeman1337/asterlink/connect"
)

func (b *b24) OrigStart(c *connect.Call, oID string) {
	b.ent[c.LID] = &entity{ID: oID, cID: c.CID, log: b.log.WithField("lid", c.LID)}
}

func (b *b24) Start(c *connect.Call) {
	b.ent[c.LID] = &entity{cID: c.CID, log: b.log.WithField("lid", c.LID)}
	e := b.ent[c.LID]

	e.mux.Lock()

	uID, ok := b.eUID[c.Ext]
	if !ok {
		uID = b.cfg.DefUID
	}

	var params struct {
		UID       int    `json:"USER_ID"`
		Phone     string `json:"PHONE_NUMBER"`
		Type      int    `json:"TYPE"`
		DID       string `json:"LINE_NUMBER"`
		CRMCreate int    `json:"CRM_CREATE"`
		Show      int    `json:"SHOW"`
	}

	params.UID = uID
	params.Phone = c.CID
	params.DID = c.DID

	if c.Dir == connect.Out {
		params.Type = 1
	} else {
		params.Type = 2
	}

	if b.cfg.CreateLeads {
		params.CRMCreate = 1
	}

	var r struct {
		Result struct {
			ID            string `json:"CALL_ID"`
			CRMEntityID   int    `json:"CRM_ENTITY_ID"`
			CRMEntityType string `json:"CRM_ENTITY_TYPE"`
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

	if r.Result.CRMEntityID != 0 {
		e.CRMEntityID = r.Result.CRMEntityID

		var method string
		switch r.Result.CRMEntityType {
		case "CONTACT":
			e.CRMEntityType = CRMEntityTypeContact

			method = "contact"
		case "COMPANY":
			e.CRMEntityType = CRMEntityTypeCompany
		case "LEAD":
			e.CRMEntityType = CRMEntityTypeLead

			method = "lead"
			if b.cfg.LeadsDeals {
				method = "deal"
			}
		default:
			r.Result.CRMEntityID = 0
			e.log.WithField("type", r.Result.CRMEntityType).Warn("Unknown CRM entity type")
		}

		if method != "" { // Query lead/contact assigned id.
			var r struct {
				Result struct {
					AssignedID string `json:"ASSIGNED_BY_ID"`
				} `json:"result"`
			}

			err := b.req("crm."+method+".get", struct {
				ID int `json:"id"`
			}{e.CRMEntityID}, &r)
			if err == nil {
				e.CRMAssignedID, _ = strconv.Atoi(r.Result.AssignedID)
			}
		}
	}

	e.log.WithField("id", e.ID).Debug("Call registred")

	e.mux.Unlock()
}

func (b *b24) Dial(c *connect.Call, ext string) {
	b.handleDial(c, ext, true)
}

func (b *b24) StopDial(c *connect.Call, ext string) {
	b.handleDial(c, ext, false)
}

func (b *b24) Answer(c *connect.Call, ext string) {
	if !b.cfg.UpdateAssigned {
		return
	}

	e, ok := b.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	if e.CRMEntityID == 0 {
		return
	}

	uID, ok := b.eUID[c.Ext]
	if !ok {
		return
	}

	if e.CRMAssignedID == uID { // No need to update assigned.
		return
	}

	var err error
	var contactLeadID int

	var method string
	switch e.CRMEntityType {
	case CRMEntityTypeLead:
		method = "lead"

		if b.cfg.LeadsDeals {
			method = "deal"
		}
	case CRMEntityTypeContact:
		method = "contact"

		var r struct {
			Result struct {
				LeadID string `json:"LEAD_ID"`
			} `json:"result"`
		}

		err = b.req("crm."+method+".get", struct {
			ID int `json:"id"`
		}{e.CRMEntityID}, &r)

		if err != nil {
			e.log.WithFields(log.Fields{"err": err}).Warn("Failed to get contact details")
		}

		contactLeadID, _ = strconv.Atoi(r.Result.LeadID)
	default:
		return
	}

	type fields struct {
		Assigned int `json:"ASSIGNED_BY_ID"`
	}

	req := struct {
		ID     int    `json:"id"`
		Fields fields `json:"fields"`
	}{
		ID:     e.CRMEntityID,
		Fields: fields{uID},
	}

	err = b.req("crm."+method+".update", req, nil)

	if err == nil && contactLeadID != 0 { // Update contact's lead as well.
		req.ID = contactLeadID

		method = "lead"
		if b.cfg.LeadsDeals {
			method = "deal"
		}
		err = b.req("crm."+method+".update", req, nil)
	}

	if err != nil {
		e.log.WithFields(log.Fields{"err": err}).Warn("Failed to change assigned contact")
	}
}

type AttachRecord struct {
	CallID     string `json:"CALL_ID"`
	FileName   string `json:"FILENAME"`
	RecURL     string `json:"RECORD_URL,omitempty"`
	B64Content string `json:"FILE_CONTENT,omitempty"`
}

func (b *b24) End(c *connect.Call, cause string) {
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

		if cause == "16" || cause == "127" { // TODO: check asterisk 18+ AST_CAUSE_INTERWORKING cause for hangup
			if c.Dir == connect.In {
				params.Status = "304" // This call was skipped
			} else {
				params.Status = "603-S" // This call was canceled
			}
		} else {
			params.Status, ok = b.causeCode[cause]
			if !ok {
				params.Status = "505" // Undefined
			}
		}
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
	if (b.cfg.RecDir != "" || b.cfg.RecUp != "") && !c.TimeAnswer.IsZero() && c.Rec != "" {
		e.log.WithFields(log.Fields{"path": c.Rec}).Debug("Attaching call record")

		r := AttachRecord{CallID: e.ID, FileName: path.Base(c.Rec)}

		if b.cfg.RecDir != "" { // Encode call record into base64.
			f, err := os.Open(b.cfg.RecDir + "/" + c.Rec)
			if err != nil {
				e.log.WithFields(log.Fields{"err": err}).Warn("Failed to open record file")
				return
			}

			recordFileB64 := &strings.Builder{}
			encoder := base64.NewEncoder(base64.StdEncoding, recordFileB64)
			_, err = io.Copy(encoder, f)
			encoder.Close()
			f.Close()

			if err != nil {
				e.log.WithFields(log.Fields{"err": err}).Warn("Failed to encode record file")
				return
			}

			r.B64Content = recordFileB64.String()
		} else if b.cfg.RecUp != "" {
			r.RecURL = b.cfg.RecUp + c.Rec
		}

		err = b.req("telephony.externalCall.attachRecord", &r, nil)
		if err != nil {
			e.log.WithFields(log.Fields{"err": err}).Warn("Failed to upload record file")
		}
	}
}

func (b *b24) handleDial(c *connect.Call, ext string, isDial bool) {
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
