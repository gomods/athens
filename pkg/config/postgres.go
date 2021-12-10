package config

// Postgres config
type Postgres struct {
	Host     string            `validate:"required" envconfig:"ATHENS_INDEX_POSTGRES_HOST"`
	Port     int               `validate:"required" envconfig:"ATHENS_INDEX_POSTGRES_PORT"`
	User     string            `validate:"required" envconfig:"ATHENS_INDEX_POSTGRES_USER"`
	Password string            `validate:"" envconfig:"ATHENS_INDEX_POSTGRES_PASSWORD"`
	Database string            `validate:"required" envconfig:"ATHENS_INDEX_POSTGRES_DATABASE"`
	Params   map[string]string `validate:"required" envconfig:"ATHENS_INDEX_POSTGRES_PARAMS"`
}
