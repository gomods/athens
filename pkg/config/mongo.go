package config

// MongoConfig specifies the properties required to use MongoDB as the storage backend.
type MongoConfig struct {
	URL                   string `envconfig:"ATHENS_MONGO_STORAGE_URL" validate:"required"`
	DefaultDBName         string `default:"athens"                     envconfig:"ATHENS_MONGO_DEFAULT_DATABASE"`
	DefaultCollectionName string `default:"modules"                    envconfig:"ATHENS_MONGO_DEFAULT_COLLECTION"`
	CertPath              string `envconfig:"ATHENS_MONGO_CERT_PATH"`
	InsecureConn          bool   `envconfig:"ATHENS_MONGO_INSECURE"`
}
