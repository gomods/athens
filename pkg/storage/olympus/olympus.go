package olympus

// ModuleStore represents a mongo backed storage backend.
type ModuleStore struct {
	url string
}

// NewStorage returns a remote Olympus store
func NewStorage(url string) *ModuleStore {
	return &ModuleStore{url: url}
}
