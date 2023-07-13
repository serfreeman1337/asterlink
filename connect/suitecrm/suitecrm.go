package suitecrm

import (
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/serfreeman1337/asterlink/connect"
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
	ent       map[string]*entity
	extUID    map[string]string
	m         sync.Mutex
	streams   streamsMap
	originate connect.OrigFunc
}

func (s *suitecrm) Init() {
	s.getExtUsers()

	if s.cfg.EndpointAddr != "" {
		http.HandleFunc("/assigned/", s.assignedHandler)

		http.Handle("/originate", s.tokenMiddleware(http.HandlerFunc(s.originateHandler)))

		s.streams = make(streamsMap)
		http.Handle("/stream", s.tokenMiddleware(http.HandlerFunc(s.streamHandler)))

		go func() {
			s.log.WithField("addr", s.cfg.EndpointAddr).Info("Enabling web server")

			srv := &http.Server{
				Addr:         s.cfg.EndpointAddr,
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
			}

			err := srv.ListenAndServe()

			if err != nil {
				s.log.Fatal(err)
			}
		}()
	}
}

func (s *suitecrm) SetOriginate(orig connect.OrigFunc) {
	s.originate = orig
}

// NewSuiteCRMConnector func
func NewSuiteCRMConnector(cfg *Config) connect.Connecter {
	s := &suitecrm{
		cfg:    cfg,
		log:    log.WithField("suite", true),
		ent:    make(map[string]*entity),
		extUID: make(map[string]string),
	}
	s.cfg.URL += "index.php?entryPoint=AsterLinkEntryPoint"

	log.Info("Using SuiteCRM Connector")

	return s
}
