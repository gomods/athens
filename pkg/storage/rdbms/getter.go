package rdbms

import (
	"bytes"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/rdbms/models"
	"io/ioutil"
	"time"
)

func (r *RDBMSModuleStore) Get(baseURL, module, vsn string) (*storage.Version, error) {
	result := models.Module{}
	query := r.conn.Where("base_url = ?", baseURL).Where("module = ?", module).Where("version = ?", vsn)
	err := query.First(&result)
	if err != nil {
		return nil, err
	}
	return &storage.Version{
		RevInfo: storage.RevInfo{
			Version: result.Version,
			Name:    result.Version,
			Short:   result.Version,
			Time:    time.Now(),
		},
		Mod: result.Mod,
		Zip: ioutil.NopCloser(bytes.NewReader(result.Zip)),
	}, nil
}
