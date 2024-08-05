package config

// AzureBlobConfig specifies the properties required to use Azure as the storage backend.
type AzureBlobConfig struct {
	AccountName               string `envconfig:"ATHENS_AZURE_ACCOUNT_NAME"                 validate:"required"`
	AccountKey                string `envconfig:"ATHENS_AZURE_ACCOUNT_KEY"`
	ManagedIdentityResourceID string `envconfig:"ATHENS_AZURE_MANAGED_IDENTITY_RESOURCE_ID"`
	CredentialScope           string `envconfig:"ATHENS_AZURE_CREDENTIAL_SCOPE"`
	ContainerName             string `envconfig:"ATHENS_AZURE_CONTAINER_NAME"               validate:"required"`
}
