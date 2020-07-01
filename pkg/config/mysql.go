package config

// MySQL config
type MySQL struct {
	Protocol string            `validate:"required" envconfig:"ATHENS_INDEX_MYSQL_PROTOCOL"`
	Host     string            `validate:"required" envconfig:"ATHENS_INDEX_MYSQL_HOST"`
	Port     int               `validate:"" envconfig:"ATHENS_INDEX_MYSQL_PORT"`
	User     string            `validate:"required" envconfig:"ATHENS_INDEX_MYSQL_USER"`
	Password string            `validate:"" envconfig:"ATHENS_INDEX_MYSQL_PASSWORD"`
	Database string            `validate:"required" envconfig:"ATHENS_INDEX_MYSQL_DATABASE"`
	Params   map[string]string `validate:"required" envconfig:"ATHENS_INDEX_MYSQL_PARAMS"`
}
