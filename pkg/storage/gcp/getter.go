package gcp

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/gomods/athens/pkg/storage"
)

// Get retrieves a module from storage as a (./pkg/storage).Version
//
// The caller is responsible for calling close on the Zip ReadCloser
func (s *Storage) Get(ctx context.Context, module, version string) (*storage.Version, error) {
	// TODO: check if module exists at version, if no - return not found
	modName := fmt.Sprintf("%s/@v/%s.%s", module, version, "mod")
	modReader, err := s.bucket.Object(modName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get new reader for mod file: %s", err)
	}
	modBytes, err := ioutil.ReadAll(modReader)
	modReader.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read bytes of mod file: %s", err)
	}

	zipName := fmt.Sprintf("%s/@v/%s.%s", module, version, "zip")
	zipReader, err := s.bucket.Object(zipName).NewReader(ctx)
	// It is up to the caller to call Close on this reader.
	// The storage.Version contains a ReadCloser for the zip.
	if err != nil {
		return nil, fmt.Errorf("could not get new reader for zip file: %s", err)
	}

	infoName := fmt.Sprintf("%s/@v/%s.%s", module, version, "info")
	infoReader, err := s.bucket.Object(infoName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get new reader for info file: %s", err)
	}
	infoBytes, err := ioutil.ReadAll(infoReader)
	infoReader.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read bytes of info file: %s", err)
	}
	return &storage.Version{Mod: modBytes, Zip: zipReader, Info: infoBytes}, nil
}
