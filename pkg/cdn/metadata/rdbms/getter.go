package rdbms

import "github.com/gomods/athens/pkg/cdn/metadata"

// Get retrieves the cdn base URL for a module
func (s *MetadataStore) Get(module string) (string, error) {
	result := metadata.CdnMetadataEntry{}
	query := s.conn.Where("module = ?", module)
	if err := query.First(&result); err != nil {
		return "", err
	}
	return result.RedirectURL, nil
}
