package storage

import (
	"io"

	"github.com/gomods/athens/pkg/observ"
)

// Getter gets module metadata and its source from underlying storage
type Getter interface {
	Info(ctx observ.ProxyContext, module, vsn string) ([]byte, error)
	GoMod(ctx observ.ProxyContext, module, vsn string) ([]byte, error)
	Zip(ctx observ.ProxyContext, module, vsn string) (io.ReadCloser, error)
}
