package main

import (
	"flag"
	"log"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/config"
)

var (
	configFile = flag.String("config_file", "../../config.toml", "The path to the config file")
)

func main() {
	flag.Parse()
	if configFile == nil {
		log.Fatal("Invalid config file path provided")
	}
	conf, err := config.ParseConfigFile(*configFile)
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
