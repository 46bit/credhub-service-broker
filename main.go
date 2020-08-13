package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/46bit/credhub-service-broker/operators"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/lager"
	"github.com/alphagov/paas-service-broker-base/broker"
)

var configFilePath string

type CredhubConfig struct {
	Prefix                 string `json:"credhub_prefix"`
	CredhubCA              string `json:"credhub_ca"`
	CredhubURL             string `json:"credhub_url"`
	CredhubUaaClientId     string `json:"credhub_uaa_client_id"`
	CredhubUaaClientSecret string `json:"credhub_uaa_client_secret"`
}

func main() {
	flag.StringVar(&configFilePath, "config", "config.json", "Location of the config file")
	flag.Parse()

	file, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("Error opening config file %s: %s\n", configFilePath, err)
	}
	defer file.Close()

	config, err := broker.NewConfig(file)
	if err != nil {
		log.Fatalf("Error validating config file: %v\n", err)
	}

	err = json.Unmarshal(config.Provider, &config)
	if err != nil {
		log.Fatalf("Error parsing configuration: %v\n", err)
	}

	var credhubConfig CredhubConfig
	err = json.Unmarshal(config.Provider, &credhubConfig)
	if err != nil {
		log.Fatalf("Error parsing configuration: %v\n", err)
	}

	logger := lager.NewLogger("credhub-service-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, config.API.LagerLogLevel))

	credhubClient, err := credhub.New(
		credhubConfig.CredhubURL,
		credhub.Auth(auth.UaaClientCredentials(
			credhubConfig.CredhubUaaClientId,
			credhubConfig.CredhubUaaClientSecret,
		)),
		credhub.CaCerts(credhubConfig.CredhubCA),
	)
	if err != nil {
		logger.Fatal("err-listening-on-port", err, lager.Data{
			"port": config.API.Port,
		})
	}

	planOperators := map[string]operators.Operator{}
	planOperators["simple"] = operators.NewSimpleOperator(credhubConfig.Prefix, credhubClient, logger)
	brokerProvider := NewBrokerProvider(planOperators, credhubConfig.Prefix, credhubClient, logger)
	serviceBroker, err := broker.New(config, brokerProvider, logger)
	if err != nil {
		logger.Fatal("err-with-new-broker", err)
	}
	brokerAPI := broker.NewAPI(serviceBroker, logger, config)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", config.API.Port))
	if err != nil {
		logger.Fatal("err-listening-on-port", err, lager.Data{
			"port": config.API.Port,
		})
	}
	http.Serve(listener, brokerAPI)
}
