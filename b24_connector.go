package main

import (
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/serfreeman1337/asterlink/connect"
	"github.com/serfreeman1337/asterlink/connect/b24"
)

func init() {
	connectors = append(connectors, newB24Connector)
}

func newB24Connector(cfgBytes []byte) (connecter connect.Connecter, err error) {
	var config struct {
		B24 b24.Config `yaml:"bitrix24"`
	}

	if err = yaml.Unmarshal(cfgBytes, &config); err != nil {
		return
	}

	if config.B24.URL == "" {
		return
	}

	if len(config.B24.FindForm) != 0 {
		for i, v := range config.B24.FindForm {
			if config.B24.FindForm[i].R, err = regexp.Compile(v.Expr); err != nil {
				return
			}
		}
		config.B24.HasFindForm = true
	}

	connecter = b24.New(&config.B24)
	return
}
