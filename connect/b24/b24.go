package b24

import (
	"crypto/tls"
	"net/http"
	"regexp"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/serfreeman1337/asterlink/connect"
)

// Config struct
type Config struct {
	IsInvalidSSL bool   `yaml:"ignore_invalid_ssl"`
	Addr         string `yaml:"webhook_endpoint_addr"`
	URL          string `yaml:"webhook_url"`
	Token        string `yaml:"webhook_originate_token"`
	RecUp        string `yaml:"rec_upload"`
	RecDir       string `yaml:"rec_dir"`
	HasFindForm  bool
	FindForm     []struct {
		R    *regexp.Regexp
		Expr string
		Repl string
	} `yaml:"search_format"`
	DefUID         int  `yaml:"default_user"`
	CreateLeads    bool `yaml:"create_leads"`
	UpdateAssigned bool `yaml:"update_assigned"`
	LeadsDeals     bool `yaml:"leads_deals"`
}

type b24 struct {
	cfg       *Config
	eUID      map[string]int
	ent       map[string]*entity
	originate connect.OrigFunc
	log       *log.Entry
	netClient *http.Client
	causeCode map[string]string

	oIDsMu sync.Mutex
	oIDs   map[string]struct{}
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

func (b *b24) SetOriginate(orig connect.OrigFunc) {
	b.originate = orig
}

// New connector
func New(cfg *Config) connect.Connecter {
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
		eUID:      make(map[string]int),
		ent:       make(map[string]*entity),
		netClient: &client,
		causeCode: map[string]string{ // https://wiki.asterisk.org/wiki/display/AST/Hangup+Cause+Mappings
			"0":  "603-S",
			"1":  "404",
			"3":  "484",
			"16": "200",
			"17": "486",
			"18": "480",
			"19": "480",
			"21": "603",
			"28": "484",
			"34": "503",
			"42": "503",
		},
		oIDs: map[string]struct{}{},
	}

	log.Info("Using Bitrix24 Connector")

	return b
}
