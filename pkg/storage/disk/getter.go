package disk

import (
	"time"

	"github.com/gomods/athens/pkg/storage"
)

func (v *storageImpl) Get(baseURL, module, vsn string) (*storage.Version, error) {
	// TODO: store the time in the saver, and parse it here
	return &storage.Version{
		RevInfo: storage.RevInfo{
			Version: vsn,
			Name:    vsn,
			Short:   vsn,
			Time:    time.Now(),
		},
	}, nil
}
