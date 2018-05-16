package main

import (
	"log"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/cmd/proxy/actions"
	mongostate "github.com/gomods/athens/pkg/proxy/state/mongo"
	"github.com/gomods/athens/pkg/storage/mongo"
)

func main() {
	app := actions.App()

	mongoURI, err := envy.MustGet("NewStorage")
	if err != nil {
		log.Fatalf("Storage uri not provided")
	}

	s := mongo.NewStorage(mongoURI)
	if err := s.Connect(); err != nil {
		log.Fatalf("Unable to connect to backing store: %v", err)
	}

	ps := mongostate.NewStateStore(mongoURI)
	if err := ps.Connect(); err != nil {
		log.Fatalf("Unable to connect to backing state store: %v", err)
	}

	w := app.Worker
	w.Register("process_module", actions.GetProcessModuleJob(s, ps))

	go actions.SyncLoop(s, ps)

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
