package gcp

import (
	"context"
	"io"
)

// Bucket defines the storage bucket TODO
type Bucket interface {
	// DeleteModule removes a module @ version from the bucket
	DeleteModule(ctx context.Context, module, version string) error
	// GetReader returns a reader for the module filename and an error if any
	GetReader(ctx context.Context, filename string) (io.ReadCloser, error)
	// GetWriter returns a new writer for the module filename
	// This will overwrite any exists file stored under the same module, version
	GetWriter(ctx context.Context, filename string) io.WriteCloser
	// ListVersions returns a slice of version strings for a module
	ListVersions(ctx context.Context, module string) ([]string, error)
	// ObjectExists returns true if the module @ version exists in the bucket
	ObjectExists(ctx context.Context, module, version string) bool
}
