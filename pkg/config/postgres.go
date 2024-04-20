package config

// Postgres config.
type Postgres struct {
	Host     string            `envconfig:"ATHENS_INDEX_POSTGRES_HOST"     validate:"required"`
	Port     int               `envconfig:"ATHENS_INDEX_POSTGRES_PORT"     validate:"required"`
	User     string            `envconfig:"ATHENS_INDEX_POSTGRES_USER"     validate:"required"`
	Password string            `envconfig:"ATHENS_INDEX_POSTGRES_PASSWORD" validate:""`
	Database string            `envconfig:"ATHENS_INDEX_POSTGRES_DATABASE" validate:"required"`
	Params   map[string]string `envconfig:"ATHENS_INDEX_POSTGRES_PARAMS"   validate:"required"`
}
