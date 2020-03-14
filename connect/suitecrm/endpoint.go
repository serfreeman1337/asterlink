package suitecrm

import (
	"fmt"
	"net/http"
	"strings"
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

func (s *suitecrm) uIDtoExt(uID string) (string, bool) {
	for k, v := range s.extUID {
		if v == uID {
			return k, true
		}
	}
	return "", false
}
