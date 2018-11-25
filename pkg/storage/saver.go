package storage

import (
	"io"

	"github.com/gomods/athens/pkg/observ"
)

// Saver saves module metadata and its source to underlying storage
type Saver interface {
	Save(ctx observ.ProxyContext, module, version string, mod []byte, zip io.Reader, info []byte) error
}
