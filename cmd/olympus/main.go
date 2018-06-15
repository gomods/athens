package main

import (
	"log"

	"github.com/gomods/athens/cmd/olympus/actions"
)

func main() {
	app := actions.App()

	eLog, err := actions.GetEventlog()
	if err != nil {
		log.Fatal(err)
	}
	storage, err := actions.GetStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Worker.Register(actions.WorkerName, actions.GetProcessPushNotificationJob(app.Worker, eLog, storage)); err != nil {
		log.Fatal(err)
	}
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
