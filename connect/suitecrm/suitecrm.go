package suitecrm

import (
	"net/http"
	"sync"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

var dirDesc = map[connect.Direction]string{connect.In: "Inbound", connect.Out: "Outbound"}

// Config struct
type Config struct {
	URL           string
	ClientID      string `yaml:"client_id"`
	ClientSecret  string `yaml:"client_secret"`
	EndpointAddr  string `yaml:"endpoint_addr"`
	EndpointToken string `yaml:"endpoint_token"`
}

type suitecrm struct {
	cfg       *Config
	log       *log.Entry
	token     string
	tokenTime time.Time
	mux       sync.Mutex
	ent       map[string]*entity
	extUID    map[string]string
	originate connect.OrigFunc
}

func (s *suitecrm) Init() {
	s.getUsers()

	if s.cfg.EndpointAddr != "" {
		http.HandleFunc("/assigned/", s.assignedHandler)
		http.HandleFunc("/originate/", s.originateHandler)
		go func() {
			s.log.WithField("addr", s.cfg.EndpointAddr).Info("Enabling web server")
			err := http.ListenAndServe(s.cfg.EndpointAddr, nil)
			if err != nil {
				s.log.Fatal(err)
			}
		}()
	}
}

// NewSuiteCRMConnector func
func NewSuiteCRMConnector(cfg *Config, originate connect.OrigFunc) connect.Connecter {
	s := &suitecrm{
		cfg: cfg,
		log: log.WithField("suite", true),
		// token:     "",
		// tokenTime: time.Now().Add(1 * time.Hour),
		ent:       make(map[string]*entity),
		extUID:    make(map[string]string),
		originate: originate,
	}
	s.cfg.URL += "Api/"

	return s
}
