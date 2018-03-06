package disk

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func (s *storageImpl) Save(baseURL, module, vsn string, mod, zip []byte) error {
	containingDir := s.versionDiskLocation(baseURL, module, vsn)
	// TODO: 777 is not the best filemode, use something better

	// make the versioned directory to hold the go.mod and the zipfile
	if err := os.MkdirAll(containingDir, 777); err != nil {
		return err
	}

	// write the go.mod file
	if err := ioutil.WriteFile(filepath.Join(containingDir, "go.mod"), mod, 777); err != nil {
		return err
	}

	// write the zipfile
	return ioutil.WriteFile(filepath.Join(containingDir, "source.zip"), zip, 777)
}
