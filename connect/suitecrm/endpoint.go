package suitecrm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/serfreeman1337/asterlink/connect"
)

type ExtKey struct{}

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

	// no relation found
	if len(e.Relationships) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var assignedID string

	for _, rel := range e.Relationships {
		if rel.AssignedID == "" {
			continue
		}

		assignedID = rel.AssignedID
		break
	}

	ext, ok := s.uIDtoExt(assignedID)
	if !ok {
		cLog.WithField("uid", assignedID).Warn("Extension not found for user id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", ext)
}

func (s *suitecrm) originateHandler(w http.ResponseWriter, r *http.Request) {
	cLog := s.log.WithField("api", "originate")

	if r.Method != "POST" {
		cLog.WithField("method", r.Method).Warn("Invalid method, only POST is allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ext := r.Context().Value(ExtKey{}).(string)
	r.ParseForm()

	var e entity

	// create new record for originated call
	err := s.createCallRecord(&connect.Call{
		CID: r.FormValue("phone"),
		Dir: connect.Out,
		Ext: ext,
	}, &e)

	if err != nil {
		cLog.Error("Failed to create call record")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.originate(ext, r.FormValue("phone"), e.ID)
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

func (s *suitecrm) tokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-AsterLink-Token")

		if r.Method == http.MethodOptions {
			return
		}

		tokenString := r.URL.Query().Get("token")

		if tokenString == "" {
			tokenString = r.Header.Get("X-Asterlink-Token")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(s.cfg.EndpointToken), nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			s.log.WithField("remote_addr", r.RemoteAddr).Warn(err.Error())
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ext, ok := s.uIDtoExt(claims["id"].(string))

			if !ok {
				http.Error(w, "Extension not found for user id", http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), ExtKey{}, ext)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
