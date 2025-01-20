package b24

import (
	"encoding/base64"
	"io"
	"os"
	"path"
	"slices"
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
	e, ok := b.ent[c.LID]
	if !ok || !e.isRegistred() {
		return
	}

	b.showsHidesMu.Lock()
	if d, ok := b.showHides[e.ID]; ok {
		d.t.Reset(0) // Send show/hide requests on Answer.
	}
	b.showsHidesMu.Unlock()

	if !b.cfg.UpdateAssigned {
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

	b.showsHidesMu.Lock() // Cancel any show/hide requests since Bitrix24 hides call card on finish anyway.
	if d, ok := b.showHides[e.ID]; ok {
		d.t.Stop()
	}
	b.showsHidesMu.Unlock()

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

	if c.Dir == connect.In { // Batch show/hide requests to improve performance with DND operators and ringalls queues.
		b.showsHidesMu.Lock()
		d, ok := b.showHides[e.ID]
		if !ok {
			t := time.AfterFunc(500*time.Millisecond, func() {
				b.showsHidesMu.Lock()

				d, ok := b.showHides[e.ID]
				if !ok {
					b.showsHidesMu.Unlock()
					return
				}

				req := func(uIDs []int, method string) {
					if len(uIDs) == 1 { // Send one uID.
						b.req(method, struct {
							ID  string `json:"CALL_ID"`
							UID int    `json:"USER_ID"`
						}{e.ID, uIDs[0]}, nil)
					} else { // Or send array of UIDs.
						b.req(method, struct {
							ID  string `json:"CALL_ID"`
							UID []int  `json:"USER_ID"`
						}{e.ID, uIDs}, nil)
					}
				}

				if len(d.showUIDs) != 0 {
					req(d.showUIDs, "telephony.externalcall.show")
				}

				if len(d.hideUIDs) != 0 {
					req(d.hideUIDs, "telephony.externalcall.hide")
				}

				delete(b.showHides, e.ID)
				b.showsHidesMu.Unlock()
			})

			d = &delayedShowHide{t, nil, nil}
			b.showHides[e.ID] = d
		} else {
			d.t.Reset(500 * time.Millisecond)
		}

		if isDial {
			if !slices.Contains(d.showUIDs, uID) { // New show UID.
				d.showUIDs = append(d.showUIDs, uID)

				if p := slices.Index(d.hideUIDs, uID); p != -1 { // Cancel its hide request.
					d.hideUIDs = slices.Delete(d.hideUIDs, p, p+1)
				}
			}
		} else {
			if p := slices.Index(d.showUIDs, uID); p != -1 { // Cancel a show request.
				d.showUIDs = slices.Delete(d.showUIDs, p, p+1)
			} else {
				d.hideUIDs = append(d.hideUIDs, uID) // New hide request.
			}
		}

		b.showsHidesMu.Unlock()
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
