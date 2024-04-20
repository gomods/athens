package config

// MySQL config.
type MySQL struct {
	Protocol string            `envconfig:"ATHENS_INDEX_MYSQL_PROTOCOL" validate:"required"`
	Host     string            `envconfig:"ATHENS_INDEX_MYSQL_HOST"     validate:"required"`
	Port     int               `envconfig:"ATHENS_INDEX_MYSQL_PORT"     validate:""`
	User     string            `envconfig:"ATHENS_INDEX_MYSQL_USER"     validate:"required"`
	Password string            `envconfig:"ATHENS_INDEX_MYSQL_PASSWORD" validate:""`
	Database string            `envconfig:"ATHENS_INDEX_MYSQL_DATABASE" validate:"required"`
	Params   map[string]string `envconfig:"ATHENS_INDEX_MYSQL_PARAMS"   validate:"required"`
}
