package actions

import (
	"context"
	"errors"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/module"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

// GetPackageDownloaderJob porcesses queue of cache misses and downloads sources from VCS
func GetPackageDownloaderJob(s storage.Backend, e eventlog.Eventlog, w worker.Worker) worker.Handler {
	return func(args worker.Args) error {
		modName, version, err := parsePackageDownloaderJobArgs(args)
		if err != nil {
			return err
		}

		// download package
		fs := afero.NewOsFs()
		f, err := module.NewGoGetFetcher(fs)
		if err != nil {
			return err
		}

		ref, err := f.Fetch(modName, version)
		if err != nil {
			return err
		}
		defer ref.Clear()

		ret, err := ref.Read()
		modBytes, infoBytes, zipFile := ret.Mod, ret.Info, ret.Zip

		// save it
		if err := s.Save(
			context.Background(),
			modName,
			version,
			modBytes,
			zipFile,
			infoBytes,
		); err != nil {
			return err
		}

		// update log
		_, err = e.Append(eventlog.Event{
			Module:  modName,
			Version: version,
			Time:    time.Now(),
			Op:      eventlog.OpAdd,
		})
		return err
	}
}

func parsePackageDownloaderJobArgs(args worker.Args) (string, string, error) {
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
