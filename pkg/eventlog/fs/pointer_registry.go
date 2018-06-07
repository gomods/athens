package fs

import (
	"encoding/gob"
	"io"
	"os"

	"github.com/gomods/athens/pkg/eventlog"
	"github.com/spf13/afero"
)

// Registry is a pointer registry for olympus server event logs
type Registry struct {
	rootDir string
	fs      afero.Fs
}

// registryData is a map[string]string used to gob encode the registry on disk
type registryData map[string]string

var registryFilename = "pointerRegistry"

// NewRegistry returns a file based implementation of a pointer registry
func NewRegistry(rootDir string, filesystem afero.Fs) *Registry {
	return &Registry{rootDir: rootDir, fs: filesystem}
}

// LookupPointer returns the pointer to the given deployment's event log
func (r *Registry) LookupPointer(deploymentID string) (string, error) {
	f, err := r.fs.OpenFile(registryFilename, os.O_RDONLY|os.O_CREATE, 0440)
	defer f.Close()
	if err != nil {
		return "", err
	}

	var data = make(registryData)

	dec := gob.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return "", err
	}

	if _, ok := data[deploymentID]; !ok {
		return "", eventlog.ErrDeploymentNotFound
	}

	return data[deploymentID], nil
}

// SetPointer both sets and updates the deployment's event log pointer
func (r *Registry) SetPointer(deploymentID, pointer string) error {
	f, err := r.fs.OpenFile(registryFilename, os.O_RDWR|os.O_CREATE, 0660)
	defer f.Close()
	if err != nil {
		return err
	}

	var data = make(registryData)

	dec := gob.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil && err != io.EOF {
		return err
	}

	data[deploymentID] = pointer

	enc := gob.NewEncoder(f)
	err = enc.Encode(&data)
	return err
}
