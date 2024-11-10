package b24

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type CRMEntityType int

const (
	CRMEntityTypeContact = iota
	CRMEntityTypeCompany
	CRMEntityTypeLead
)

type entity struct {
	ID            string
	cID           string
	CRMEntityID   int
	CRMEntityType CRMEntityType
	log           *log.Entry
	mux           sync.Mutex
}

func (e *entity) isRegistred() bool {
	e.mux.Lock()

	registred := e.ID != ""

	e.mux.Unlock()

	return registred
}
