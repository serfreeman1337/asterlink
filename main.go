package main

import (
	"io/ioutil"
	"os"

	"github.com/serfreeman1337/asterlink/connect"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type newConnectorFunc func([]byte) (connect.Connecter, error)

var connectors []newConnectorFunc

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	log.Info("AsterLink")

	var amiCfg AmiConfig
	var connector connect.Connecter

	if cfgBytes, err := ioutil.ReadFile("conf.yml"); err == nil {
		var config struct {
			LogLevel log.Level `yaml:"log_level"`
		}

		if err := yaml.Unmarshal(cfgBytes, &config); err != nil {
			log.Fatal(err)
		}

		if err := yaml.Unmarshal(cfgBytes, &amiCfg); err != nil {
			log.Fatal(err)
		}

		log.WithField("level", config.LogLevel).Info("Setting log level")
		log.SetLevel(config.LogLevel)

		for _, getConnector := range connectors {
			if connector, err = getConnector(cfgBytes); err != nil {
				log.Fatal(err)
			}

			if connector != nil {
				break
			}
		}
	} else {
		log.Fatal(err)
	}

	if connector == nil {
		log.Warn("No connector selected")
		connector = connect.NewDummyConnector()
	}

	ami, err := NewAmiConnector(&amiCfg)

	if err != nil {
		log.Fatal(err)
	}

	connector.SetOriginate(ami.Originate)
	connector.Init()

	ami.SetConnector(connector)
	ami.Init()

	r := make(chan bool)
	<-r
}
