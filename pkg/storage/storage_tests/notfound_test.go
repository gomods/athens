package storagetest

import (
	"github.com/gomods/athens/pkg/storage"
)

func (d *StorageTests) TestNotFound() {
	r := d.Require()

	for _, store := range d.storages {
		_, err := store.Get("some", "unknown")

		r.Equal(true, storage.IsNotFoundError(err), "Invalid error type for %s: %#v", store, err)
	}
}
