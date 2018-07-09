package storage

// StorageTest is common interface which each storage needs to implement
type StorageTest interface {
	TestNotFound()
}
