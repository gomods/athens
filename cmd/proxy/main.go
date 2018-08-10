package main

import (
	"log"

	"github.com/gomods/athens/cmd/proxy/actions"
	"github.com/gomods/athens/pkg/module"
)

func main() {
	mf := module.NewFilter()
	app, err := actions.App(mf)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
