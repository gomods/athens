package config

// Index is the config for various index storage backends
type Index struct {
	MySQL    *MySQL
	Postgres *Postgres
}
