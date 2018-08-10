package config

import (
	"github.com/spf13/viper"
)

type StorageConfig struct {
	CDN   *CDNConfig   `validate:""`
	Disk  *DiskConfig  `validate:""`
	GCP   *GCPConfig   `validate:""`
	Minio *MinioConfig `validate:""`
	Mongo *MongoConfig `validate:""`
	RDBMS *RDBMSConfig `validate:""`
}

// This method might get a lot less verbose if https://github.com/spf13/viper/issues/307 gets resolved
// At the moment it's not possible to modify a sub viper and propagate the values back up
func setStorageDefaults(v *viper.Viper) {
	if !v.IsSet("Storage") {
		return
	}

	globalTimeout := v.GetInt("TimeoutSeconds")

	if v.IsSet("Storage.CDN") {
		v.BindEnv("Storage.CDN.Endpoint", "CDN_ENDPOINT")
		v.SetDefault("Storage.CDN.TimeoutSeconds", globalTimeout)
	}

	if v.IsSet("Storage.Disk") {
		v.BindEnv("Storage.Disk.RootPath", "ATHENS_DISK_STORAGE_ROOT")
		v.SetDefault("Storage.Disk.TimeoutSeconds", globalTimeout)
	}

	if v.IsSet("Storage.GCP") {
		v.BindEnv("Storage.GCP.ProjectID", "GOOGLE_CLOUD_PROJECT")
		v.BindEnv("Storage.GCP.Bucket", "ATHENS_STORAGE_GCP_BUCKET")
		v.SetDefault("Storage.GCP.TimeoutSeconds", globalTimeout)
	}

	if v.IsSet("Storage.Minio") {

		v.BindEnv("Storage.Minio.Endpoint", "ATHENS_MINIO_ENDPOINT")
		v.BindEnv("Storage.Minio.Key", "ATHENS_MINIO_ACCESS_KEY_ID")
		v.BindEnv("Storage.Minio.Secret", "ATHENS_MINIO_SECRET_ACCESS_KEY")

		v.BindEnv("Storage.Minio.Bucket", "ATHENS_MINIO_BUCKET_NAME")
		v.SetDefault("Storage.Minio.Bucket", "gomods")

		v.BindEnv("Storage.Minio.EnableSSL", "ATHENS_MINIO_USE_SSL")
		v.SetDefault("Storage.Minio.EnableSSL", true)

		v.SetDefault("Storage.Minio.TimeoutSeconds", globalTimeout)
	}

	if v.IsSet("Storage.Mongo") {

		v.BindEnv("Storage.Mongo.URL", "ATHENS_MONGO_STORAGE_URL")
		v.BindEnv("Storage.Mongo.User", "MONGO_USER")
		v.BindEnv("Storage.Mongo.Password", "MONGO_PASSWORD")

		v.SetDefault("Storage.Mongo.TimeoutSeconds", globalTimeout)
	}

	if v.IsSet("Storage.RDBMS") {
		v.BindEnv("Storage.RDBMS.Name", "ATHENS_RDBMS_STORAGE_NAME")
		v.SetDefault("Storage.RDBMS.TimeoutSeconds", globalTimeout)
	}

}
