package main

import (
	"fmt"
	"log"

	"github.com/gomods/athens/cmd/proxy/actions"
)

func main() {
	store, err := actions.GetStorage()
	if err != nil {
		err = fmt.Errorf("error getting storage configuration (%s)", err)
		log.Fatal(err)
	}
	app, err := actions.App(store)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
