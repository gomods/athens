package config

// AzureBlobConf specifies properties required for using azureblob as a storage backend
type AzureBlobConf struct {
	AccountName   string
	AccountKey    string
	ContainerName string
}
