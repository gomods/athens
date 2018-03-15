package mongo

import (
	"errors"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/gomods/athens/pkg/storage"
)

func (s *storageImpl) List(baseURL, module string) ([]string, error) {
	c := s.s.DB(s.d).C(s.c)
	result := make([]*storage.Module, 0)
	err := c.Find(bson.M{"BaseUrl": baseURL, "Name": module}).All(result)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = errors.New("module not found")
		}
	}

	versions := make([]string, len(result))
	for i, r := range result {
		versions[i] = r.Version
	}

	return versions, nil
}
