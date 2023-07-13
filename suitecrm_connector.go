package main

import (
	"gopkg.in/yaml.v3"

	"github.com/serfreeman1337/asterlink/connect"
	"github.com/serfreeman1337/asterlink/connect/suitecrm"
)

func init() {
	connectors = append(connectors, newSuiteCRMConnector)
}

func newSuiteCRMConnector(cfgBytes []byte) (connecter connect.Connecter, err error) {
	var config struct {
		SuiteCRM suitecrm.Config `yaml:"suitecrm"`
	}

	if err = yaml.Unmarshal(cfgBytes, &config); err != nil {
		return
	}

	if config.SuiteCRM.URL == "" {
		return
	}

	connecter = suitecrm.New(&config.SuiteCRM)
	return
}
