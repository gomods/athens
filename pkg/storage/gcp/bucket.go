package gcp

import (
	"io"

	"github.com/gomods/athens/pkg/observ"
)

// Bucket provides file operations for a Google Cloud Storage resource.
type Bucket interface {
	// Delete removes the file
	Delete(ctx observ.ProxyContext, path string) error
	// Open returns a reader for a path and any error
	Open(ctx observ.ProxyContext, path string) (io.ReadCloser, error)
	// Write returns a new writer for a path
	// This writer will overwrite any existing file stored at the same path
	Write(ctx observ.ProxyContext, path string) io.WriteCloser
	// List returns a slice of paths for a prefix and any error
	List(ctx observ.ProxyContext, prefix string) ([]string, error)
	// Exists returns true if the file exists
	Exists(ctx observ.ProxyContext, path string) (bool, error)
}
