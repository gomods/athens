package config

type DiskConfig struct {
	RootPath       string `validate:"required"`
	TimeoutSeconds int
}
