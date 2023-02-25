package config

// S3Config specifies the properties required to use S3 as the storage backend.
type S3Config struct {
	Region                             string `validate:"required" envconfig:"AWS_REGION"`
	Key                                string `envconfig:"AWS_ACCESS_KEY_ID"`
	Secret                             string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	Token                              string `envconfig:"AWS_SESSION_TOKEN"`
	Bucket                             string `validate:"required" envconfig:"ATHENS_S3_BUCKET_NAME"`
	UseDefaultConfiguration            bool   `envconfig:"AWS_USE_DEFAULT_CONFIGURATION"`
	ForcePathStyle                     bool   `envconfig:"AWS_FORCE_PATH_STYLE"`
	CredentialsEndpoint                string `envconfig:"AWS_CREDENTIALS_ENDPOINT"`
	AwsContainerCredentialsRelativeURI string `envconfig:"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"`
	Endpoint                           string `envconfig:"AWS_ENDPOINT"`
}
