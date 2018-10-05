package config

// S3Config specifies the properties required to use S3 as the storage backend
type S3Config struct {
	TimeoutConf
	Region         string `validate:"required" envconfig:"AWS_REGION"`
	Key            string `validate:"required" envconfig:"AWS_ACCESS_KEY_ID"`
	Secret         string `validate:"required" envconfig:"AWS_SECRET_ACCESS_KEY"`
	Token          string `envconfig:"AWS_SESSION_TOKEN"`
	Bucket         string `validate:"required" envconfig:"ATHENS_S3_BUCKET_NAME"`
	Endpoint       string
	DisableSSL     bool
	ForcePathStyle bool
}
