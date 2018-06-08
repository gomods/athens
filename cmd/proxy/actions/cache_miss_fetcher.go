package actions

import (
	"errors"
	"io/ioutil"
	"log"
	"sync"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/storage"
	olympusStore "github.com/gomods/athens/pkg/storage/olympus"
)

const (
	// OlympusGlobalEndpoint is a default olympus DNS address
	OlympusGlobalEndpoint = "olympus.gomods.io"
)

var (
	currentOlympusEndpoint = OlympusGlobalEndpoint
	endpointLock           sync.Mutex
)

// GetProcessCacheMissJob porcesses queue of cache misses and downloads sources from active Olympus
func GetProcessCacheMissJob(s storage.Backend, w worker.Worker) worker.Handler {
	return func(args worker.Args) (err error) {
		module, version, err := parseArgs(args)
		if err != nil {
			return err
		}

		if s.Exists(module, version) {
			return nil
		}

		// get module info
		v, err := getModuleInfo(module, version)
		if err != nil {
			process(module, version, args, w)
			return err
		}

		zip, err := ioutil.ReadAll(v.Zip)
		if err != nil {
			process(module, version, args, w)
			return err
		}

		if err = s.Save(module, version, v.Mod, zip, v.Info); err != nil {
			process(module, version, args, w)
		}

		return err
	}
}

func parseArgs(args worker.Args) (string, string, error) {
	module, ok := args[workerModuleKey].(string)
	if !ok {
		return "", "", errors.New("module name not specified")
	}

	version, ok := args[workerVersionKey].(string)
	if !ok {
		return "", "", errors.New("version not specified")
	}

	return module, version, nil
}

func getModuleInfo(module, version string) (*storage.Version, error) {
	olympusEndpoint := getCurrentOlympus()
	os := olympusStore.NewStorage(olympusEndpoint)

	v, err := os.Get(module, version)
	if err != nil {
		if redirectErr, ok := err.(*eventlog.ErrUseNewOlympus); ok {
			olympusEndpoint := redirectErr.Endpoint
			updateCurrentOlympus(olympusEndpoint)
		}

		return nil, err
	}

	return v, nil
}

// process pushes pull job into the queue to be processed asynchonously
func process(module, version string, args worker.Args, w worker.Worker) error {
	// decrementing avoids endless loop of entries with missing trycount
	trycount, _ := args[workerTryCountKey].(int)
	if trycount <= 0 {
		log.Printf("Max trycount for %s %s reached\n", module, version)
	}

	return w.Perform(worker.Job{
		Queue:   workerQueue,
		Handler: FetcherWorkerName,
		Args: worker.Args{
			workerModuleKey:   module,
			workerVersionKey:  version,
			workerTryCountKey: trycount - 1,
		},
	})
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
