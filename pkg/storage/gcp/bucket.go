package gcp

import (
	"context"
	"io"
)

// Bucket provides file operations for a Google Cloud Storage resource.
type Bucket interface {
	// Delete removes the file module/@v/version.extension
	Delete(ctx context.Context, path string) error
	// Open returns a reader for module/@v/version.extension and any error
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	// Write returns a new writer for module/@v/version.extension
	// This writer will overwrite any existing file stored at the same path
	Write(ctx context.Context, path string) io.WriteCloser
	// ListVersions returns a slice of versions for a module and any error
	List(ctx context.Context, prefix string) ([]string, error)
	// Exists returns true if the module @ version exists
	Exists(ctx context.Context, path string) bool
}
