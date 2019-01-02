package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

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
	var conf *config.Config

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		log.Print("Config file not found - using default settings")
		conf = config.CreateDefault()
	} else {
		conf, err = config.ParseConfigFile(*configFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	handler, err := actions.App(conf)
	if err != nil {
		log.Fatal(err)
	}

	cert, key, err := conf.TLSCertFiles()
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:    conf.Port,
		Handler: handler,
	}
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Starting application at port %v", conf.Port)
	if cert != "" && key != "" {
		err = srv.ListenAndServeTLS(conf.TLSCertFile, conf.TLSKeyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-idleConnsClosed
}
