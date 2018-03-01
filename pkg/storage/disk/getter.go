package disk

import (
	"github.com/gomods/athens/pkg/storage"
)

type Getter struct{}

func (v *listerSaverImpl) Get(baseURL, module, vsn string) (*storage.Version, error) {
	return nil, nil
}
