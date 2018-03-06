package disk

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func (s *storageImpl) Save(baseURL, module, vsn string, mod, zip []byte) error {
	p := &payload{
		baseURL:     baseURL,
		module:      module,
		version:     vsn,
		moduleBytes: mod,
		zipBytes:    zip,
	}
	containingDir := p.diskLocation()
	// TODO: 777 is not the best filemode, use something better

	// make the versioned directory to hold the go.mod and the zipfile
	if err := os.MkdirAll(containingDir, 777); err != nil {
		return err
	}

	// write the go.mod file
	if err := ioutil.WriteFile(filepath.Join(p.diskLocation(), "go.mod"), p.moduleBytes, 777); err != nil {
		return err
	}

	// write the zipfile
	return ioutil.WriteFile(filepath.Join(p.diskLocation(), "source.zip"), p.zipBytes, 777)
}
