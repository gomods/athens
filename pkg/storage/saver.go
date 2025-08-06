package storage

import (
	"context"
	"io"
)

// Saver saves module metadata and its source to underlying storage.
type Saver interface {
	// Save saves the module metadata and its source to the storage.
	//
	// The caller MAY call zipMD5 with a nil value if the checksum is not available.
	// The storage implementation MAY use the zipMD5 to verify the integrity of the zip file.
	Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, zipMD5, info []byte) error
}
