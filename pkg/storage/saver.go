package storage

import (
	"github.com/gobuffalo/buffalo"
)

// Saver saves module metadata and its source to underlying storage
type Saver interface {
	Save(c buffalo.Context, module, version string, mod, zip, info []byte) error
}
