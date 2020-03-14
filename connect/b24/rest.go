package b24

import (
	"bytes"
	"encoding/json"
	"errors"
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
	return nil, errors.New("Contact not found")
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
