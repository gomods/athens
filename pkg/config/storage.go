package config

// Storage provides configs for various storage backends.
type Storage struct {
	Disk      *DiskStorage
	GCP       *GCPStorage
	Minio     *MinioStorage
	Mongo     *MongoStorage
	S3        *S3Storage
	AzureBlob *AzureBlobStorage
	External  *ExternalStorage
}

// DiskStorage specifies the properties required to use Disk as the storage backend.
type DiskStorage struct {
	RootPath string `envconfig:"ATHENS_DISK_STORAGE_ROOT" validate:"required"`
}

// GCPStorage specifies the properties required to use GCP as the storage backend.
type GCPStorage struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT"`
	Bucket    string `envconfig:"ATHENS_STORAGE_GCP_BUCKET"   validate:"required"`
	JSONKey   string `envconfig:"ATHENS_STORAGE_GCP_JSON_KEY"`
}

// MinioStorage specifies the properties required to use Minio or DigitalOcean Spaces
// as the storage backend.
type MinioStorage struct {
	Endpoint  string `envconfig:"ATHENS_MINIO_ENDPOINT"          validate:"required"`
	Key       string `envconfig:"ATHENS_MINIO_ACCESS_KEY_ID"     validate:"required"`
	Secret    string `envconfig:"ATHENS_MINIO_SECRET_ACCESS_KEY" validate:"required"`
	Bucket    string `envconfig:"ATHENS_MINIO_BUCKET_NAME"       validate:"required"`
	Region    string `envconfig:"ATHENS_MINIO_REGION"`
	EnableSSL bool   `envconfig:"ATHENS_MINIO_USE_SSL"`
}

// MongoStorage specifies the properties required to use MongoDB as the storage backend.
type MongoStorage struct {
	URL                   string `envconfig:"ATHENS_MONGO_STORAGE_URL" validate:"required"`
	DefaultDBName         string `default:"athens"                     envconfig:"ATHENS_MONGO_DEFAULT_DATABASE"`
	DefaultCollectionName string `default:"modules"                    envconfig:"ATHENS_MONGO_DEFAULT_COLLECTION"`
	CertPath              string `envconfig:"ATHENS_MONGO_CERT_PATH"`
	InsecureConn          bool   `envconfig:"ATHENS_MONGO_INSECURE"`
}

// S3Storage specifies the properties required to use S3 as the storage backend.
type S3Storage struct {
	Region                             string `envconfig:"AWS_REGION"                             validate:"required"`
	Key                                string `envconfig:"AWS_ACCESS_KEY_ID"`
	Secret                             string `envconfig:"AWS_SECRET_ACCESS_KEY"`
	Token                              string `envconfig:"AWS_SESSION_TOKEN"`
	Bucket                             string `envconfig:"ATHENS_S3_BUCKET_NAME"                  validate:"required"`
	UseDefaultConfiguration            bool   `envconfig:"AWS_USE_DEFAULT_CONFIGURATION"`
	ForcePathStyle                     bool   `envconfig:"AWS_FORCE_PATH_STYLE"`
	CredentialsEndpoint                string `envconfig:"AWS_CREDENTIALS_ENDPOINT"`
	AwsContainerCredentialsRelativeURI string `envconfig:"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"`
	Endpoint                           string `envconfig:"AWS_ENDPOINT"`
}

// AzureBlobStorage specifies the properties required to use Azure as the storage backend.
type AzureBlobStorage struct {
	AccountName               string `envconfig:"ATHENS_AZURE_ACCOUNT_NAME"                 validate:"required"`
	AccountKey                string `envconfig:"ATHENS_AZURE_ACCOUNT_KEY"`
	ManagedIdentityResourceID string `envconfig:"ATHENS_AZURE_MANAGED_IDENTITY_RESOURCE_ID"`
	CredentialScope           string `envconfig:"ATHENS_AZURE_CREDENTIAL_SCOPE"`
	ContainerName             string `envconfig:"ATHENS_AZURE_CONTAINER_NAME"               validate:"required"`
}

// ExternalStorage specifies configuration for an external http storage.
type ExternalStorage struct {
	URL string `envconfig:"ATHENS_EXTERNAL_STORAGE_URL" validate:"required"`
}
