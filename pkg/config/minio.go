package config

// MinioConfig specifies the properties required to use Minio as the storage backend
type MinioConfig struct {
	Endpoint  string `validate:"required" envconfig:"ATHENS_MINIO_ENDPOINT"`
	Key       string `validate:"required" envconfig:"ATHENS_MINIO_ACCESS_KEY_ID"`
	Secret    string `validate:"required" envconfig:"ATHENS_MINIO_SECRET_ACCESS_KEY"`
	Timeout   int    `validate:"required"`
	Bucket    string `validate:"required" envconfig:"ATHENS_MINIO_BUCKET_NAME" default:"gomods"`
	EnableSSL bool   `envconfig:"ATHENS_MINIO_USE_SSL" default:"true"`
}
