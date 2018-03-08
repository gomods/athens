package disk

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	baseURL = "testbaseurl"
	module  = "testmodule"
	version = "v1.0.0"
)

var (
	mod = []byte("123")
	zip = []byte("456")
)

type DiskTests struct {
	suite.Suite
	storage Storage
	rootDir string
}

func (d *DiskTests) SetupTest() {
	r, err := ioutil.TempDir("", "athens-disk-tests")
	d.Require().NoError(err)
	d.storage = NewStorage(r)
	d.rootDir = r
}

func (d *DiskTests) TearDownTest() {
	d.Require().NoError(os.RemoveAll(d.rootDir))
}

func TestDiskStorage(t *testing.T) {
	suite.Run(t, new(DiskTests))
}
