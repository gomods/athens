package storage

// Backend is a complete storage backend (i.e. file system, database) implementation - a reader, saver and deleter
type Backend interface {
	Reader
	Saver
	Deleter
}
