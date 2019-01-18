package config

// HTTPConfig specifies the properties required to use HTTP as the storage backend
type HTTPConfig struct {
	BaseURL  string `validate:"required" envconfig:"ATHENS_HTTP_BASE_URL"`
	Username string `envconfig:"ATHENS_HTTP_USERNAME"`
	Password string `envconfig:"ATHENS_HTTP_PASSWORD"`
}
