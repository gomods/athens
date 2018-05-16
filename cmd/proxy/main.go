package main

import (
	"log"

	"github.com/gomods/athens/cmd/proxy/actions"
)

func main() {
	app := actions.App()

	w := app.Worker
	w.Register("process_module", processModuleJob)

	go actions.SyncLoop()

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
