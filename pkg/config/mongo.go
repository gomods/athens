package config

type MongoConfig struct {
	URL            string `validate:"required"`
	User           string
	Password       string
	TimeoutSeconds int `validate:"required"`
}
