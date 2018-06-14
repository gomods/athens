package rdbms

import (
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/rdbms/models"

	"github.com/gobuffalo/buffalo"
)

// asserts that ModuleStore implements storage.Saver
var _ storage.Saver = &ModuleStore{}

// Save stores a module in rdbms storage.
func (r *ModuleStore) Save(_ buffalo.Context, module, version string, mod, zip, info []byte) error {
	m := &models.Module{
		Module:  module,
		Version: version,
		Mod:     mod,
		Zip:     zip,
		Info:    info,
	}

	return r.conn.Create(m)
}
