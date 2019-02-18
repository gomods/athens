package storage

import "io"

// Zip represents zip file of a specific version and it's size
type Zip struct {
	Zip  io.ReadCloser
	Size int64
}
