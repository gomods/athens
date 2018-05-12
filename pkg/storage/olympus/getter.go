package olympus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gomods/athens/pkg/storage"
)

// Get a specific version of a module
func (s *ModuleStore) Get(module, vsn string) (*storage.Version, error) {
	// TODO: fetch from endpoint

	modURI := fmt.Sprintf("%s/%s/@v/%s.mod", s.url, module, vsn)
	zipURI := fmt.Sprintf("%s/%s/@v/%s.zip", s.url, module, vsn)

	// fetch mod file
	var mod []byte
	client := http.Client{Timeout: 180 * time.Second}
	modResp, err := client.Get(modURI)
	if err != nil {
		return nil, err
	}
	defer modResp.Body.Close()

	mod, err = ioutil.ReadAll(modResp.Body)
	if err != nil {
		return nil, err
	}

	// fetch source file
	var zip []byte
	zipResp, err := client.Get(zipURI)
	if err != nil {
		return nil, err
	}
	defer zipResp.Body.Close()

	zip, err = ioutil.ReadAll(zipResp.Body)
	if err != nil {
		return nil, err
	}

	return &storage.Version{
		RevInfo: storage.RevInfo{
			Version: vsn,
			Name:    module,
			Short:   module,
			Time:    time.Now(),
		},
		Mod: mod,
		Zip: ioutil.NopCloser(bytes.NewReader(zip)),
	}, nil
}
