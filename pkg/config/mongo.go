package config

// MongoConfig specifies the properties required to use MongoDB as the storage backend
type MongoConfig struct {
	URL          string `validate:"required" envconfig:"ATHENS_MONGO_STORAGE_URL"`
	CertPath     string `envconfig:"ATHENS_MONGO_CERT_PATH"`
	InsecureConn bool   `envconfig:"ATHENS_MONGO_INSECURE"`
}
