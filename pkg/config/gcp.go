package config

// GCPConfig specifies the properties required to use GCP as the storage backend.
type GCPConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT"`
	Bucket    string `envconfig:"ATHENS_STORAGE_GCP_BUCKET"   validate:"required"`
	JSONKey   string `envconfig:"ATHENS_STORAGE_GCP_JSON_KEY"`
}
