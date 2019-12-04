package config

// GCPConfig specifies the properties required to use GCP as the storage backend
type GCPConfig struct {
	ProjectID      string `envconfig:"GOOGLE_CLOUD_PROJECT"`
	Bucket         string `validate:"required" envconfig:"ATHENS_STORAGE_GCP_BUCKET"`
	ServiceAccount string `envconfig:"ATHENS_STORAGE_GCP_SERVICE_ACCOUNT"`
	// NOTE: JSONKey is deprecated in favor of ServiceAccount above.
	// We've left it here here for backward compatibility.
	// If both ServiceAccount and JSONKey are set, ServiceAccount will take precedence
	JSONKey string `envconfig:"ATHENS_STORAGE_GCP_JSON_KEY"`
}
