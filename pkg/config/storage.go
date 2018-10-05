package config

import validator "gopkg.in/go-playground/validator.v9"

// StorageConfig provides configs for various storage backends
type StorageConfig struct {
	CDN   *CDNConfig
	Disk  *DiskConfig
	GCP   *GCPConfig
	Minio *MinioConfig
	Mongo *MongoConfig
	S3    *S3Config
}

func setStorageTimeouts(s *StorageConfig, defaultTimeout int) {
	if s == nil {
		return
	}
	if s.CDN != nil && s.CDN.Timeout == 0 {
		s.CDN.Timeout = defaultTimeout
	}
	if s.GCP != nil && s.GCP.Timeout == 0 {
		s.GCP.Timeout = defaultTimeout
	}
	if s.Minio != nil && s.Minio.Timeout == 0 {
		s.Minio.Timeout = defaultTimeout
	}
	if s.Mongo != nil && s.Mongo.Timeout == 0 {
		s.Mongo.Timeout = defaultTimeout
	}
	if s.S3 != nil && s.S3.Timeout == 0 {
		s.S3.Timeout = defaultTimeout
	}
}

// envconfig initializes *all* struct pointers, even if there are no corresponding defaults or env variables
// deleteInvalidStorageConfigs prunes all such invalid configurations
func deleteInvalidStorageConfigs(s *StorageConfig) {
	validate := validator.New()

	if s.CDN != nil {
		if err := validate.Struct(s.CDN); err != nil {
			s.CDN = nil
		}
	}

	if s.Disk != nil {
		if err := validate.Struct(s.Disk); err != nil {
			s.Disk = nil
		}
	}

	if s.GCP != nil {
		if err := validate.Struct(s.GCP); err != nil {
			s.GCP = nil
		}
	}

	if s.Minio != nil {
		if err := validate.Struct(s.Minio); err != nil {
			s.Minio = nil
		}
	}

	if s.Mongo != nil {
		if err := validate.Struct(s.Mongo); err != nil {
			s.Mongo = nil
		}
	}

	if s.S3 != nil {
		if err := validate.Struct(s.S3); err != nil {
			s.S3 = nil
		}
	}
}
