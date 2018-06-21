package storage

type noOpBackend struct {
	backend Backend
}

// NoOpBackend wraps storage backend with Connect functionality
func NoOpBackend(b Backend) Backend {
	return noOpBackend{backend: b}
}

func (n noOpBackend) Exists(module, version string) bool {
	return n.backend.Exists(module, version)
}

func (n noOpBackend) Get(module, vsn string) (*Version, error) {
	return n.backend.Get(module, vsn)
}
func (n noOpBackend) List(module string) ([]string, error) {
	return n.backend.List(module)
}
func (n noOpBackend) Save(module, version string, mod, zip, info []byte) error {
	return n.backend.Save(module, version, mod, zip, info)
}
func (n noOpBackend) Delete(module, version string) error {
	return n.backend.Delete(module, version)
}
