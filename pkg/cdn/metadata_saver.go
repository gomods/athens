package cdn

// MetadataSaver saves metadata about the module/version's location in a CDN.
// Returns a non-nil error if there were any issues saving this information
type MetadataSaver interface {
	// Save saves the module/version information to metadata storage
	Save(moduleName, version, cdnLocation string) error
}
