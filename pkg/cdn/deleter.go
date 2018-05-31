package cdn

// Deleter deletes a module from the CDN and its metadata from metadata storage
type Deleter interface {
	// Delete removes module/version from the CDN and its metadata storage.
	// Returns ErrNotFound if the module/version isn't found, and another
	// non-nil error on any other error encountered
	Delete(module string, version string) error
}
