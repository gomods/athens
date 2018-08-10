package config

type RDBMSStorage interface {
	RdbmsName() (string, error)
}

type RDBMSConfig struct {
	Name           string `validate:"required"`
	TimeoutSeconds int
}
