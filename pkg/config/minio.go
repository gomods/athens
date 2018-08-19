package config

// MinioConfig specifies the properties required to use Minio as the storage backend
type MinioConfig struct {
	Endpoint  string `validate:"required" envconfig:"ATHENS_MINIO_ENDPOINT"`
	Key       string `validate:"required" envconfig:"ATHENS_MINIO_ACCESS_KEY_ID"`
	Secret    string `validate:"required" envconfig:"ATHENS_MINIO_SECRET_ACCESS_KEY"`
	Timeout   int    `validate:"required"`
	Bucket    string `validate:"required" envconfig:"ATHENS_MINIO_BUCKET_NAME"`
	EnableSSL *bool  `envconfig:"ATHENS_MINIO_USE_SSL"`
}

func setMinioDefaults(m *MinioConfig, timeout int) *MinioConfig {
	if m == nil {
		return nil
	}
	overrideDefaultStr(&m.Bucket, "gomods")
	overrideDefaultInt(&m.Timeout, timeout)
	if m.EnableSSL == nil {
		t := true
		m.EnableSSL = &t
	}
	return m
}
