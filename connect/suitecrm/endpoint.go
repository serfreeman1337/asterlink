package suitecrm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/serfreeman1337/asterlink/connect"
)

func (s *suitecrm) assignedHandler(w http.ResponseWriter, r *http.Request) {
	cLog := s.log.WithField("api", "assigned")
	req := strings.Split(r.RequestURI, "/")[1:]

	if len(req[1]) == 0 {
		cLog.WithField("path", r.RequestURI).Warn("Incorrect RequestURI")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, ok := s.ent[req[1]]
	if !ok || !e.isRegistred() {
		cLog.WithField("lid", req[1]).Warn("Call not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// search for contact
	_, assigned, err := s.findContact(e.cID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ext, ok := s.uIDtoExt(assigned)
	if !ok {
		cLog.WithField("uid", assigned).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (s *suitecrm) originateHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	cLog := s.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("method", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.ParseForm()

	if r.FormValue("token") != s.cfg.EndpointToken {
		cLog.WithField("remote_addr", r.RemoteAddr).Warn("Invalid endpoint token")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ext, ok := s.uIDtoExt(r.FormValue("user"))
	if !ok {
		cLog.WithField("uid", r.FormValue("user")).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// create new record for originated call
	id, err := s.createCallRecord(&connect.Call{
		CID: r.FormValue("phone"),
		Dir: connect.Out,
		Ext: ext,
	})

	if err != nil {
		cLog.Error("Failed to create call record")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.originate(ext, r.FormValue("phone"), id)
	w.WriteHeader(http.StatusOK)
}

func (s *suitecrm) uIDtoExt(uID string) (string, bool) {
	for k, v := range s.extUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
