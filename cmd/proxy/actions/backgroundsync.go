package actions

import (
	"errors"
	"io/ioutil"
	"sync"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/eventlog/olympus"
	proxystate "github.com/gomods/athens/pkg/proxy/state"
	"github.com/gomods/athens/pkg/storage"
	olympusStore "github.com/gomods/athens/pkg/storage/olympus"
)

const (
	// OlympusGlobalEndpoint is a default olympus DNS address
	OlympusGlobalEndpoint = "olympus.gomods.io"
)

var (
	currentOlympusEndpoint = ""
	endpointLock           sync.Mutex
)

// GetProcessModuleJob porcesses job from a queue and downloads missing module
func GetProcessModuleJob(s storage.Backend, ps proxystate.Store, w worker.Worker) worker.Handler {
	return func(args worker.Args) (err error) {
		currentOlympusEndpoint := getCurrentOlympus()
		jobOlympusEndpoint, ok := args[workerEndpointKey].(string)
		if !ok {
			return errors.New("olympus endpoint not provided")
		}

		// if processing job associated with older state, skip
		if currentOlympusEndpoint != jobOlympusEndpoint {
			return nil
		}

		event, ok := args[workerEventKey].(eventlog.Event)
		if !ok {
			return errors.New("event to process not provided")
		}

		if s.Exists(event.Module, event.Version) {
			return nil
		}

		// get module info
		os := olympusStore.NewStorage(jobOlympusEndpoint)
		version, err := os.Get(event.Module, event.Version)
		if err != nil {
			process(jobOlympusEndpoint, event, w)
			return err
		}

		zip, err := ioutil.ReadAll(version.Zip)
		if err != nil {
			process(jobOlympusEndpoint, event, w)
			return err
		}

		err = s.Save(event.Module, event.Version, version.Mod, zip)
		if err != nil {
			process(jobOlympusEndpoint, event, w)
		}

		return err
	}
}

// SyncLoop is synchronization background job meant to run in a goroutine
// pulling event log from olympus
func SyncLoop(s storage.Backend, ps proxystate.Store, w worker.Worker) {
	olympusEndpoint, sequenceID := getLoopState(ps)
	updateCurrentOlympus(olympusEndpoint)

	for {
		select {
		case <-time.After(30 * time.Second):
			ee, err := getEventLog(olympusEndpoint, sequenceID)

			if err != nil {
				// on redirect from global to deployment update state,
				if redirectErr, ok := err.(*eventlog.ErrUseNewOlympus); ok {
					olympusEndpoint, sequenceID = redirectErr.Endpoint, ""
					ps.Set(olympusEndpoint, sequenceID)
					updateCurrentOlympus(olympusEndpoint)
					continue
				}
				// on another error reset state
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
		return OlympusGlobalEndpoint, ""
	}

	return state.OlympusEndpoint, state.SequenceID
}

func updateCurrentOlympus(o string) {
	endpointLock.Lock()
	defer endpointLock.Unlock()
	currentOlympusEndpoint = o
}

func getCurrentOlympus() string {
	endpointLock.Lock()
	defer endpointLock.Unlock()
	return currentOlympusEndpoint
}
