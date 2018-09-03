package main

import (
	"log"

	"github.com/gomods/athens/cmd/olympus/actions"
	"github.com/gomods/athens/pkg/config"
)

const (
	configPath = "../../config.toml"
)

func main() {
	conf, err := config.ParseConfigFile(configPath)
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
