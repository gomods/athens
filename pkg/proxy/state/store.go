package state

// Store is interface covering backend implementation of the state
type Store interface {
	Setter
	Getter
}
