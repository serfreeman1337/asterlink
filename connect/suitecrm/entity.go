package suitecrm

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
	if e.ID != "" {
		return true
	}

	e.mux.Lock()
	e.mux.Unlock()

	if e.ID == "" {
		return false
	}

	return true
}
