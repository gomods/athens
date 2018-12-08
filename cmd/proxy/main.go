package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gobuffalo/buffalo/servers"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/build"
	"github.com/gomods/athens/pkg/config"
)

var (
	configFile = flag.String("config_file", filepath.Join("..", "..", "config.dev.toml"), "The path to the config file")
	version    = flag.Bool("version", false, "Print version information and exit")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(build.String())
		os.Exit(0)
	}
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

	cert, key, err := conf.TLSCertFiles()
	if err != nil {
		log.Fatal(err)
	}

	var srv servers.Server
	if cert != "" && key != "" {
		srv = servers.WrapTLS(&http.Server{}, conf.TLSCertFile, conf.TLSKeyFile)
	} else {
		srv = servers.Wrap(&http.Server{})
	}

	if err := app.Serve(srv); err != nil {
		log.Fatal(err)
	}
}
