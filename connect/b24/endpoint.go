package b24

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (b *b24) apiOriginateHandler(w http.ResponseWriter, r *http.Request) {
	cLog := b.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("method", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseForm()

	if r.FormValue("auth[application_token]") != b.cfg.Token {
		cLog.WithField("remote_addr", r.RemoteAddr).Warn("Invalid webhook token")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	uID, _ := strconv.Atoi(r.FormValue("data[USER_ID]"))
	ext, ok := b.uIDtoExt(uID)
	if !ok {
		cLog.WithField("uid", uID).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b.originate(ext, r.FormValue("data[PHONE_NUMBER_INTERNATIONAL]"), r.FormValue("data[CALL_ID]"))
	w.WriteHeader(http.StatusOK)
}

func (b *b24) apiAssignedHandler(w http.ResponseWriter, r *http.Request) {
	cLog := b.log.WithField("api", "assigned")
	req := strings.Split(r.RequestURI, "/")[1:]

	if len(req[1]) == 0 {
		cLog.WithField("path", r.RequestURI).Warn("Incorrect RequestURI")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, ok := b.ent[req[1]]
	if !ok || !e.isRegistred() {
		cLog.WithField("lid", req[1]).Warn("Call not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// search for contact
	contact, err := b.findContact(e.cID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext, ok := b.uIDtoExt(contact.Assigned)
	if !ok {
		cLog.WithField("uid", contact.Assigned).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (b *b24) uIDtoExt(uID int) (string, bool) {
	for k, v := range b.eUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}
