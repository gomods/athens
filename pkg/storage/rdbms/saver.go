package rdbms

import (
	"github.com/gomods/athens/pkg/storage/rdbms/models"
)

func (r *RDBMSModuleStore) Save(baseURL, module, version string, mod, zip []byte) error {
	m := &models.Module{
		BaseURL: baseURL,
		Module:  module,
		Version: version,
		Mod:     mod,
		Zip:     zip,
	}

	return r.conn.Create(m)
}
