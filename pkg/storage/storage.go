package storage

// Backend is a complete storage backend implementation - a lister, reader and saver
type Backend interface {
	Lister
	Getter
	Saver
}
