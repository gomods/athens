package olympus

// Save stores a module in olympus.
// This actually does not store anything just reports cache miss
func (s *ModuleStore) Save(module, version string, _, _ []byte) error {
	return nil
}
