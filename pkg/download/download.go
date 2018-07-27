package download

import (
	"context"
	"io"

	"github.com/gomods/athens/pkg/storage"
)

// Protocol is the download protocol which mirrors
// the http requests that cmd/go makes to the proxy.
type Protocol interface {
	List(ctx context.Context, module string) ([]string, error)
	Info(ctx context.Context, module, version string) (*storage.RevInfo, error)
	Latest(ctx context.Context, module string) (*storage.RevInfo, error)
	GoMod(ctx context.Context, module, version string) ([]byte, error)
	Zip(ctx context.Context, module, version string) (io.ReadCloser, error)
}
