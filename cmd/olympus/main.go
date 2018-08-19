package main

import (
	"log"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/cmd/olympus/actions"
	"github.com/gomods/athens/pkg/config"
)

const (
	configFile = "../../config.test.toml"
)

func main() {
	app, err := setupApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

func setupApp() (*buffalo.App, error) {

	conf, err := config.ParseConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	return actions.App(conf)
}
