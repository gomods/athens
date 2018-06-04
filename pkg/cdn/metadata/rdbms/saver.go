package rdbms

import "github.com/gomods/athens/pkg/cdn/metadata/rdbms/models"

// Save saves the module and it's cdn base URL.
func (s *MetadataStore) Save(module, redirectURL string) error {
	r := models.CdnMetadataEntry{Module: module, RedirectURL: redirectURL}
	return s.conn.Create(&r)
}
