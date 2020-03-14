package suitecrm

import (
	"sync"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

var dirDesc = map[connect.Direction]string{connect.In: "Inbound", connect.Out: "Outbound"}

// Config struct
type Config struct {
	URL          string
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type suitecrm struct {
	cfg       *Config
	log       *log.Entry
	token     string
	tokenTime time.Time
	mux       sync.Mutex
	ent       map[string]*entity
	extUID    map[string]string
}

func (s *suitecrm) Init() {
	s.getUsers()
}

// NewSuiteCRMConnector func
func NewSuiteCRMConnector(cfg *Config, originate connect.OrigFunc) connect.Connecter {
	s := &suitecrm{
		cfg: cfg,
		log: log.WithField("suite", true),
		// token:     "",
		// tokenTime: time.Now().Add(1 * time.Hour),
		ent:    make(map[string]*entity),
		extUID: make(map[string]string),
	}
	s.cfg.URL += "Api/"

	return s
}
