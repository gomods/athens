package config

type GCPConfig struct {
	ProjectID      string
	Bucket         string `validate:"required"`
	TimeoutSeconds int    `validate:"required"`
}
