package config

// Index is the config for various index storage backends.
type Index struct {
	MySQL    *MySQLIndex
	Postgres *PostgresIndex
}

// MySQLIndex config for using MySQL as an indexer.
type MySQLIndex struct {
	Protocol string            `envconfig:"ATHENS_INDEX_MYSQL_PROTOCOL" validate:"required"`
	Host     string            `envconfig:"ATHENS_INDEX_MYSQL_HOST"     validate:"required"`
	Port     int               `envconfig:"ATHENS_INDEX_MYSQL_PORT"     validate:""`
	User     string            `envconfig:"ATHENS_INDEX_MYSQL_USER"     validate:"required"`
	Password string            `envconfig:"ATHENS_INDEX_MYSQL_PASSWORD" validate:""`
	Database string            `envconfig:"ATHENS_INDEX_MYSQL_DATABASE" validate:"required"`
	Params   map[string]string `envconfig:"ATHENS_INDEX_MYSQL_PARAMS"   validate:"required"`
}

// PostgresIndex config for using Postgres as an indexer.
type PostgresIndex struct {
	Host     string            `envconfig:"ATHENS_INDEX_POSTGRES_HOST"     validate:"required"`
	Port     int               `envconfig:"ATHENS_INDEX_POSTGRES_PORT"     validate:"required"`
	User     string            `envconfig:"ATHENS_INDEX_POSTGRES_USER"     validate:"required"`
	Password string            `envconfig:"ATHENS_INDEX_POSTGRES_PASSWORD" validate:""`
	Database string            `envconfig:"ATHENS_INDEX_POSTGRES_DATABASE" validate:"required"`
	Params   map[string]string `envconfig:"ATHENS_INDEX_POSTGRES_PARAMS"   validate:"required"`
}
