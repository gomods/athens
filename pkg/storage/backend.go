package storage

// Backend is a complete storage backend (i.e. file system, database) implementation - a lister, reader and saver.
type Backend interface {
	Lister
	Getter
	Saver
	Deleter
}
# Env override: ATHENS_STORAGE_TYPE
StorageType = "external"

[Storage]
    [Storage.External]
        # Env override: ATHENS_EXTERNAL_STORAGE_URL
        URL = "http://localhost:9090"
