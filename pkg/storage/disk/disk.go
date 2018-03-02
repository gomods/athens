package disk

import (
	"path/filepath"

	"github.com/gomods/athens/pkg/storage"
)

// Storage is the only interface defined by the disk storage. Use
// NewStorage to create one of these. Everything is all in one
// because it all has to share the same tree
type Storage interface {
	storage.Lister
	storage.Getter
	storage.Saver
}

type storageImpl struct {
	rootDir string
}

// NewStorage returns a new ListerSaver implementation that stores
// everything under rootDir
func NewStorage(rootDir string) Storage {
	return &storageImpl{rootDir: rootDir}

}

type payload struct {
	root        string
	baseURL     string
	module      string
	version     string
	moduleBytes []byte
	zipBytes    []byte
}

func (p *payload) diskLocation() string {
	return filepath.Join(p.root, p.baseURL, p.module, p.version)
}
