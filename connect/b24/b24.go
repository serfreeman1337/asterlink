package b24

import (
	"crypto/tls"
	"net/http"
	"regexp"
	"time"

	"github.com/serfreeman1337/asterlink/connect"
	log "github.com/sirupsen/logrus"
)

// Config struct
type Config struct {
	IsInvalidSSL bool   `yaml:"ignore_invalid_ssl"`
	Addr         string `yaml:"webhook_endpoint_addr"`
	URL          string `yaml:"webhook_url"`
	Token        string `yaml:"webhook_originate_token"`
	RecUp        string `yaml:"rec_upload"`
	HasFindForm  bool
	FindForm     []struct {
		R    *regexp.Regexp
		Expr string
		Repl string
	} `yaml:"search_format"`
	DefUID int `yaml:"default_user"`
}

type b24 struct {
	cfg       *Config
	eUID      map[string]int
	ent       map[string]*entity
	originate connect.OrigFunc
	log       *log.Entry
	netClient *http.Client
}

func (b *b24) Init() {
	b.updateUsers()

	if b.cfg.Addr != "" {
		http.HandleFunc("/assigned/", b.apiAssignedHandler)
		http.HandleFunc("/originate/", b.apiOriginateHandler)
		go func() {
			b.log.WithField("addr", b.cfg.Addr).Info("Enabling web server")
			err := http.ListenAndServe(b.cfg.Addr, nil)
			if err != nil {
				b.log.Fatal(err)
			}
		}()
	}
}

// NewB24Connector connector
func NewB24Connector(cfg *Config, originate connect.OrigFunc) connect.Connecter {
	client := http.Client{
		Timeout: time.Second * 30,
	}

	if cfg.IsInvalidSSL {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
		log.Info("Ignore invalid self-signed certificates")
	}

	b := &b24{
		cfg:       cfg,
		log:       log.WithField("b24", true),
		originate: originate,
		eUID:      make(map[string]int),
		ent:       make(map[string]*entity),
		netClient: &client,
	}

	return b
}
