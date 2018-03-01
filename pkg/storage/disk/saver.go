package disk

import (
	"io/ioutil"
	"path/filepath"
)

func (s *listerSaverImpl) Save(baseURL, module, vsn string, mod, zip []byte) error {
	p := &payload{
		baseURL:     baseURL,
		module:      module,
		version:     vsn,
		moduleBytes: mod,
		zipBytes:    zip,
	}
	return p.writeGoMod()
}

func (p *payload) writeGoMod() error {
	// TODO: 777 is not the best filemode, use something better
	return ioutil.WriteFile(filepath.Join(p.diskLocation(), "go.mod"), p.moduleBytes, 777)
}

func (p *payload) writeZip() error {
	// TODO: 777 is not the best filemode, use something better
	return ioutil.WriteFile(filepath.Join(p.diskLocation(), "source.zip"), p.zipBytes, 777)
}
