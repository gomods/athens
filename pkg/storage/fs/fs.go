package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/spf13/afero"
)

type storageImpl struct {
	rootDir    string
	filesystem afero.Fs
}

func (s *storageImpl) moduleLocation(module string) string {
	return filepath.Join(s.rootDir, module)
}

func (s *storageImpl) versionLocation(module, version string) string {
	return filepath.Join(s.moduleLocation(module), version)

}

// NewStorage returns a new ListerSaver implementation that stores
// everything under rootDir
// If the root directory does not exist an error is returned
func NewStorage(rootDir string, filesystem afero.Fs) (storage.Backend, error) {
	const op errors.Op = "fs.NewStorage"
	exists, err := afero.Exists(filesystem, rootDir)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not check if root directory `%s` exists: %s", rootDir, err))
	}
	if !exists {
		return nil, errors.E(op, fmt.Errorf("root directory `%s` does not exist", rootDir))
	}
	return &storageImpl{rootDir: rootDir, filesystem: filesystem}, nil
}

func (s *storageImpl) Clear() error {
	if err := s.filesystem.RemoveAll(s.rootDir); err != nil {
		return err
	}
	return s.filesystem.Mkdir(s.rootDir, os.ModeDir|os.ModePerm)
}
