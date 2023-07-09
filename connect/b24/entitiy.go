package b24

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type entity struct {
	ID  string
	cID string
	log *log.Entry
	mux sync.Mutex
}

func (e *entity) isRegistred() bool {
	e.mux.Lock()

	registred := e.ID != ""

	e.mux.Unlock()

	return registred
}
