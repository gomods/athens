package mongo

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/gomods/athens/pkg/storage"
)

// Save stores a module in mongo storage.
func (s *ModuleStore) Save(_ context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	z, err := ioutil.ReadAll(zip)
	if err != nil {
		return err
	}
	m := &storage.Module{
		Module:  module,
		Version: version,
		Mod:     mod,
		Zip:     z,
		Info:    info,
	}

	c := s.s.DB(s.d).C(s.c)
	return c.Insert(m)
}
