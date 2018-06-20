package storage

type noOpConnectedBackend struct {
	backend Backend
}

// NoOpBackend wraps storage backend with Connect functionality
func NoOpBackend(b Backend) Backend {
	return noOpConnectedBackend{backend: b}
}

func (n noOpConnectedBackend) Exists(module, version string) bool {
	return n.backend.Exists(module, version)
}

func (n noOpConnectedBackend) Get(module, vsn string) (*Version, error) {
	return n.backend.Get(module, vsn)
}
func (n noOpConnectedBackend) List(module string) ([]string, error) {
	return n.backend.List(module)
}
func (n noOpConnectedBackend) Save(module, version string, mod, zip, info []byte) error {
	return n.backend.Save(module, version, mod, zip, info)
}
func (n noOpConnectedBackend) Delete(module, version string) error {
	return n.backend.Delete(module, version)
}
