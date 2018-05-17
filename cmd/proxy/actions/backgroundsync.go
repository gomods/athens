package actions

import (
	"errors"
	"io/ioutil"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/olympus"
	proxystate "github.com/gomods/athens/pkg/proxy/state"
	"github.com/gomods/athens/pkg/storage"
	olympusStore "github.com/gomods/athens/pkg/storage/olympus"
)

// GetProcessModuleJob porcesses job from a queue and downloads missing module
func GetProcessModuleJob(s storage.Backend, ps proxystate.Store) worker.Handler {
	return func(args worker.Args) error {
		olympusEndpoint, ok := args[workerEndpointKey].(string)
		if !ok {
			return errors.New("olympus endpoint not provided")
		}
		event, ok := args[workerEventKey].(eventlog.Event)
		if !ok {
			return errors.New("event to process not provided")
		}

		if s.Exists(event.Module, event.Version) {
			return nil
		}

		// get module info
		os := olympusStore.NewStorage(olympusEndpoint)
		version, err := os.Get(event.Module, event.Version)
		if err != nil {
			return err
		}

		zip, err := ioutil.ReadAll(version.Zip)
		if err != nil {
			return err
		}

		return s.Save(event.Module, event.Version, version.Mod, zip)
	}
}

// SyncLoop is synchronization background job meant to run in a goroutine
// pulling event log from olympus
func SyncLoop(s storage.Backend, ps proxystate.Store, w worker.Worker) {
	olympusEndpoint, sequenceID := getLoopState(ps)

	for {
		select {
		case <-time.After(30 * time.Second):
			ee, err := getEventLog(olympusEndpoint, sequenceID)

			if err != nil {
				ps.Clear()
				olympusEndpoint, sequenceID = getLoopState(ps)
				continue
			}

			for _, e := range ee {
				err = process(olympusEndpoint, e, w)
				if err != nil {
					break
				}
			}

			if len(ee) > 0 {
				lastID := ee[len(ee)-1].ID
				ps.Set(olympusEndpoint, lastID)
			}
		}
	}
}

// Process pushes pull job into the queue to be processed asynchonously
func process(olympusEndpoint string, event eventlog.Event, w worker.Worker) error {
	return w.Perform(worker.Job{
		Queue:   workerQueue,
		Handler: workerName,
		Args: worker.Args{
			workerEndpointKey: olympusEndpoint,
			workerEventKey:    event,
		},
	})
}

func getEventLog(olympusEndpoint string, sequenceID string) ([]eventlog.Event, error) {
	olympusReader := olympus.NewLog(olympusEndpoint)

	if sequenceID == "" {
		return olympusReader.Read()
	}

	return olympusReader.ReadFrom(sequenceID)
}

func getLoopState(ps proxystate.Store) (olympusEndpoint string, sequenceID string) {
	// try env overrides
	olympusEndpoint = envy.Get("PROXY_OLYMPUS_ENDPOINT", "")
	sequenceID = envy.Get("PROXY_SEQUENCE_ID", "")

	state, err := ps.Get()
	if err != nil {
		return "olympus.gomods.io", ""
	}

	return state.OlympusEndpoint, state.SequenceID
}
