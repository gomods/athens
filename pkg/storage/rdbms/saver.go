package rdbms

import (
	"context"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/rdbms/models"
)

// asserts that ModuleStore implements storage.Saver
var _ storage.Saver = &ModuleStore{}

// Save stores a module in rdbms storage.
func (r *ModuleStore) Save(_ context.Context, module, version string, mod, zip, info []byte) error {
	m := &models.Module{
		Module:  module,
		Version: version,
		Mod:     mod,
		Zip:     zip,
		Info:    info,
	}

	return r.conn.Create(m)
}
