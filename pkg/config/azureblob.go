package config

// AzureBlobConfig specifies properties required for using azureblob
// as a storage backend
type AzureBlobConfig struct {
	AccountName   string
	AccountKey    string
	ContainerName string
}
