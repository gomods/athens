package config

// RDBMSConfig specifies the properties required to use an RDBMS as the storage backend
type RDBMSConfig struct {
	Name    string `validate:"required" envconfig:"ATHENS_RDBMS_STORAGE_NAME"`
	Timeout int
}
