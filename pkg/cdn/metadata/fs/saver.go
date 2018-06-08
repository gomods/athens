package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/gobuffalo/uuid"
	"github.com/gomods/athens/pkg/cdn/metadata"
)

// Save saves the module and it's cdn base URL as a json file.
// it returns ErrExist if the metadata file already exists
func (s *fsStore) Save(module, redirectURL string) error {
	d := filepath.Join(s.rootDir, module)
	if err := s.filesystem.MkdirAll(d, os.ModePerm); err != nil {
		return err
	}
	p := filepath.Join(d, metadataFileName)
	// os.O_CREATE|os.O_EXCL so we get an err if the file exists
	f, err := s.filesystem.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	m := &metadata.CDNMetadataEntry{
		ID:          id,
		Module:      module,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	enc := json.NewEncoder(f)
	return enc.Encode(m)
}
