package config

type MinioConfig struct {
	Endpoint       string `validate:"required"`
	Key            string `validate:"required"`
	Secret         string `validate:"required"`
	TimeoutSeconds int    `validate:"required"`
	Bucket         string `validate:"required"`
	EnableSSL      bool   `validate:"required"`
}
