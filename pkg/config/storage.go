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
