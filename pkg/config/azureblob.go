package config

// AzureBlobConfig specifies the properties required to use Azure as the storage backend
type AzureBlobConfig struct {
	TimeoutConf
	AccountName   string `validate:"required" envconfig:"ATHENS_AZURE_ACCOUNT_NAME"`
	AccountKey    string `validate:"required" envconfig:"ATHENS_AZURE_ACCOUNT_KEY"`
	ContainerName string `validate:"required" envconfig:"ATHENS_AZURE_CONTAINER_NAME"`
}
