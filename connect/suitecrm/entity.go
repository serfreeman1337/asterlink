package suitecrm

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/serfreeman1337/asterlink/connect"
)

type entity struct {
	ID string `json:"id"`

	Dir           connect.Direction   `json:"dir"`
	DID           string              `json:"did"`
	CID           string              `json:"cid"`
	IsAnswered    bool                `json:"answered"`
	TimeStamp     time.Time           `json:"time"`
	Relationships map[string]relation `json:"relations"`
	exts          sync.Map

	log *log.Entry
	mux sync.Mutex
}

func (e *entity) isRegistred() bool {
	e.mux.Lock()

	registred := e.ID != ""

	e.mux.Unlock()

	return registred
}
