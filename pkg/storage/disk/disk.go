package disk

import (
	"io/ioutil"
	"path/filepath"

	"github.com/gomods/athens/pkg/storage"
)

// ListerSaver is the only interface defined by the disk storage. Use
// NewListerSaver to create one of these. Everything is all in one
// because it all has to share the same tree
type ListerSaver interface {
	storage.Lister
	storage.Getter
	storage.Saver
}

type listerSaverImpl struct {
	rootDir string
}

// NewListerSaver returns a new ListerSaver implementation that stores
// everything under rootDir
func NewListerSaver(rootDir string) ListerSaver {
	return &listerSaverImpl{rootDir: rootDir}

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
