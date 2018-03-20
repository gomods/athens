package rdbms

import (
	"github.com/gomods/athens/pkg/storage/rdbms/models"
)

func (r *RDBMSModuleStore) List(baseURL, module string) ([]string, error) {
	result := make([]models.Module, 0)
	err := r.conn.Where("base_url = ?", baseURL).Where("module = ?", module).All(&result)
	if err != nil {
		return nil, err
	}

	versions := make([]string, len(result))
	for i := range result {
		versions[i] = result[i].Version
	}

	return versions, nil
}
