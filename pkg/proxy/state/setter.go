package state

// Setter sets the state of the proxy
type Setter interface {
	Set(endpoint, sequenceID string) error
	Clear() error
}
