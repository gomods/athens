package disk

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/gomods/athens/pkg/storage"
)

func (v *storageImpl) Get(baseURL, module, vsn string) (*storage.Version, error) {
	versionedPath := v.versionDiskLocation(baseURL, module, vsn)

	mod, err := ioutil.ReadFile(filepath.Join(versionedPath, "go.mod"))
	if err != nil {
		return nil, err
	}

	src, err := ioutil.ReadFile(filepath.Join(versionedPath, "source.zip"))
	if err != nil {
		return nil, err
	}

	// TODO: store the time in the saver, and parse it here
	return &storage.Version{
		RevInfo: storage.RevInfo{
			Version: vsn,
			Name:    vsn,
			Short:   vsn,
			Time:    time.Now(),
		},
		Mod: mod,
		Zip: src,
	}, nil
}
