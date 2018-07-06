package gcp

// List implements the (./pkg/storage).Lister interface
// It returns a list of versions, if any, for a given module
func (s *Storage) List(module string) ([]string, error) {
	return nil, nil
}
