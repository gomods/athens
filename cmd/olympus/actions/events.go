package actions

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/worker"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/payloads"
)

func eventlogHandler(r eventlog.Reader) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		seqID := c.Param("sequence_id")

		var events []eventlog.Event
		var err error
		if seqID == "" {
			events, err = r.Read()
		} else {
			events, err = r.ReadFrom(seqID)
		}
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, renderEng.JSON(events))
	}
}

func cachemissHandler(l eventlog.Appender) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		cm := &payloads.Module{}
		if err := c.Bind(cm); err != nil {
			return err
		}
		e := eventlog.Event{Module: cm.Name, Version: cm.Version, Time: time.Now(), Op: eventlog.OpAdd}
		id, err := l.Append(e)
		if err != nil {
			return err
		}
		e.ID = id
		return c.Render(http.StatusOK, renderEng.JSON(e))
	}
}

func pushNotificationHandler(w worker.Worker) func(c buffalo.Context) error {
	return func(c buffalo.Context) error {
		p := &payloads.PushNotification{}
		if err := c.Bind(p); err != nil {
			return err
		}
		process(p, w)
	}
}

func process(pn payloads.PushNotification, w worker.Worker) error {
	p, err := json.Marshal(pn)
	if err != nil {
		return err
	}

	return w.Perform(worker.Job{
		Queue:   workerQueue,
		Handler: WorkerName,
		Args: worker.Args{
			workerPushNotificationKey: string(p),
		},
	})
}
