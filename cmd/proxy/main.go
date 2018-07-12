package main

import (
	"fmt"
	"log"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/gocraft-work-adapter"
	"github.com/gocraft/work"
	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/storage"
)

func main() {
	app, err := actions.App()
	if err != nil {
		log.Fatal(err)
	}

	s, err := getLocalStorage()
	if err != nil {
		log.Fatal(err)
	}

	if w, ok := app.Worker.(*gwa.Adapter); ok {
		registerWithOptions(w, s)
	} else {
		w := app.Worker
		registerNoOptions(w, s)
	}

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

func registerWithOptions(w *gwa.Adapter, s storage.Backend) {
	opts := work.JobOptions{
		SkipDead: true,
		MaxFails: 5,
	}

	mf := module.NewFilter()
	if err := w.RegisterWithOptions(actions.FetcherWorkerName, opts, actions.GetProcessCacheMissJob(s, w, mf)); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWithOptions(actions.ReporterWorkerName, opts, actions.GetCacheMissReporterJob(w, mf)); err != nil {
		log.Fatal(err)
	}
}

func registerNoOptions(w worker.Worker, s storage.Backend) {
	mf := module.NewFilter()
	if err := w.Register(actions.FetcherWorkerName, actions.GetProcessCacheMissJob(s, w, mf)); err != nil {
		log.Fatal(err)
	}

	if err := w.Register(actions.ReporterWorkerName, actions.GetCacheMissReporterJob(w, mf)); err != nil {
		log.Fatal(err)
	}
}
