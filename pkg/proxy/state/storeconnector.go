package state

import "github.com/gomods/athens/pkg/proxy"

// StoreConnector is a regular state storage backend with Connect functionality
type StoreConnector interface {
	Store
	Connect() error
}

type noOpStoreConnector struct {
	s Store
}

// NoOpStoreConnector wraps storage backend with Connect functionality
func NoOpStoreConnector(s Store) StoreConnector {
	return noOpStoreConnector{s: s}
}

func (n noOpStoreConnector) Connect() error {
	return nil
}

func (n noOpStoreConnector) Set(endpoint, sequenceID string) error {
	return n.s.Set(endpoint, sequenceID)
}

func (n noOpStoreConnector) Clear() error {
	return n.s.Clear()
}

func (n noOpStoreConnector) Get() (proxy.State, error) {
	return n.s.Get()
}
