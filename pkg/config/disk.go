package config

// DiskConfig specifies the properties required to use Disk as the storage backend.
type DiskConfig struct {
	RootPath string `envconfig:"ATHENS_DISK_STORAGE_ROOT" validate:"required"`
}
