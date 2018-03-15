package mongo

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/storage"
)

func (s *storageImpl) Get(baseURL, module, vsn string) (*storage.Version, error) {
	c := s.s.DB(s.d).C(s.c)
	result := &storage.Module{}
	err := c.Find(bson.M{"BaseUrl": baseURL, "Name": module, "Version": vsn}).One(result)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = errors.New("module not found")
		}
	}
	return &storage.Version{
		RevInfo: storage.RevInfo{
			Version: result.Version,
			Name:    result.Version,
			Short:   result.Version,
			Time:    time.Now(),
		},
		Mod: result.Mod,
		Zip: ioutil.NopCloser(bytes.NewReader(result.Zip)),
	}, nil
}
