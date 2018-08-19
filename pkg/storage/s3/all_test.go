package s3

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

const (
	testConfigFile = "../../../config.test.toml"
)

type S3Tests struct {
	suite.Suite
	uploader *s3UploaderMock
	storage  *Storage
}

func getConf(t *testing.T) *config.Config {
	absPath, err := filepath.Abs(testConfigFile)
	if err != nil {
		t.Errorf("Unable to construct absolute path to test config file")
	}
	conf, err := config.ParseConfigFile(absPath)
	if err != nil {
		t.Errorf("Unable to parse config file %s", err.Error())
	}
	return conf
}

func Test_ActionSuite(t *testing.T) {
	conf := getConf(t)
	if conf.Storage.CDN == nil {
		t.Errorf("Invalid CDN Configuration provided")
	}
	uploaderMock := newUploaderMock()
	storage, err := NewWithUploader("test", uploaderMock, conf.Storage.CDN)
	if err != nil {
		t.Error(err)
	}

	suite.Run(t, &S3Tests{uploader: uploaderMock, storage: storage})
}

// Verify returns error if S3 state differs from expected one
func Verify(um *s3UploaderMock, value map[string][]byte) error {
	um.lock.Lock()
	defer um.lock.Unlock()

	expectedLength := len(value)
	actualLength := len(um.db)
	if expectedLength != actualLength {
		return fmt.Errorf("Length does not match. Expected: %d. Actual: %d", expectedLength, actualLength)
	}

	for k, v := range value {
		actual, ok := um.db[k]
		if !ok {
			return fmt.Errorf("Missing element %s", k)
		}

		if !sliceEqualCheck(v, actual) {
			return fmt.Errorf("Value for key %s does not match. Expected: %v, Actual: %v", k, v, actual)
		}
	}

	return nil
}

func sliceEqualCheck(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
