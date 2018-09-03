package main

import (
	"log"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/config"
)

const (
	configFile = "../../config.toml"
)

func main() {
	conf, err := config.ParseConfigFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	app, err := actions.App(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
