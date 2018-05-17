package main

import (
	"fmt"
	"log"

	"github.com/gomods/athens/pkg/proxy/state"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/storage"
)

func main() {
	app := actions.App()

	s, err := getLocalStorage()
	if err != nil {
		log.Fatal(err)
	}

	ps, err := getStateStorage()
	if err != nil {
		log.Fatal(err)
	}

	w := app.Worker
	if err := w.Register("process_module", actions.GetProcessModuleJob(s, ps)); err != nil {
		log.Fatal(err)
	}

	go actions.SyncLoop(s, ps, w)

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}

func getLocalStorage() (storage.Backend, error) {
	s, err := actions.GetStorage()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve backing store: %v", err)
	}

	if err := s.Connect(); err != nil {
		return nil, fmt.Errorf("Unable to connect to backing store: %v", err)
	}

	return s, nil
}

func getStateStorage() (state.Store, error) {
	ps, err := actions.GetStateStorage()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve backing state store: %v", err)
	}

	if err := ps.Connect(); err != nil {
		return nil, fmt.Errorf("Unable to connect to backing state store: %v", err)
	}

	return ps, nil
}
