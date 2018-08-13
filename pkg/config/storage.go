package config

// StorageConfig provides configs for various storage backends
type StorageConfig struct {
	CDN   *CDNConfig   `validate:""`
	Disk  *DiskConfig  `validate:""`
	GCP   *GCPConfig   `validate:""`
	Minio *MinioConfig `validate:""`
	Mongo *MongoConfig `validate:""`
	RDBMS *RDBMSConfig `validate:""`
}

func setDefaultTimeouts(s *StorageConfig, defaultTimeout int) {
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
	if s.RDBMS != nil && s.RDBMS.Timeout == 0 {
		s.RDBMS.Timeout = defaultTimeout
	}
}

// envconfig initializes *all* struct pointers, even if there are no corresponding defaults or env variables
// deleteInvalidStorageConfigs prunes all such invalid configurations
func deleteInvalidStorageConfigs(s *StorageConfig) {
	validate := getValidator()

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

	if s.RDBMS != nil {
		if err := validate.Struct(s.RDBMS); err != nil {
			s.RDBMS = nil
		}
	}
}
