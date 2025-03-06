package b24

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type contactInfo struct {
	Type     string `json:"CRM_ENTITY_TYPE"`
	ID       int    `json:"CRM_ENTITY_ID"`
	Assigned int    `json:"ASSIGNED_BY_ID"`
}

func (b *b24) findContact(phone string) (*contactInfo, error) {
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
			Result []contactInfo
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
	return nil, errors.New("contact not found")
}

func (b *b24) updateUsers() {
	nTotal := 1
	nRet := 0
	clear(b.eUID)

ret:
	for nTotal > nRet {
		var r struct {
			Result []struct {
				ID    int    `json:"ID,string"`
				Phone string `json:"UF_PHONE_INNER"`
			}
			Total int
		}

		err := b.req("user.get", map[string]interface{}{
			"sort":   "UF_PHONE_INNER", // desc sorting for UF_PHONE_INNER should bring users without assigned extensions to the end
			"order":  "DESC",
			"filter": map[string]string{"USER_TYPE": "employee"},
			"start":  nRet,
		}, &r)

		if err != nil {
			b.log.Error("Failed to update users list")
			return
		}

		if r.Total == 0 || len(r.Result) == 0 {
			break
		}

		nTotal = r.Total
		nRet += len(r.Result)

		for _, v := range r.Result {
			if v.Phone == "" { // there should be no more users with assigned extensions
				break ret
			}

			b.eUID[v.Phone] = v.ID
		}
	}

	b.log.WithField("users", b.eUID).Info("User list updated")
}

func (b *b24) req(method string, params interface{}, result interface{}) (err error) {
	data := &bytes.Buffer{}
	err = json.NewEncoder(data).Encode(params)
	if err != nil {
		b.log.Error(err)
		return
	}

	if method == "telephony.externalCall.attachRecord" {
		params.(*AttachRecord).B64Content = "< REMOVED >"
	}

	b.log.WithFields(log.Fields{"method": method}).Trace(params)

	res, err := b.netClient.Post(b.cfg.URL+method+"/", "application/json", data)
	if err != nil {
		b.log.Error(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = errors.New(res.Status)
		stuff, _ := io.ReadAll(res.Body)
		b.log.WithFields(log.Fields{"msg": string(stuff)}).Error(err)
		return
	}

	if result != nil {
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			b.log.Error(err)
			return
		}

		b.log.WithFields(log.Fields{"method": method}).Trace(result)
	} else {
		io.Copy(io.Discard, res.Body)
	}

	return
}
