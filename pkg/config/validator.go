package config

// Validator can validate a config struct. If you implement this,
// validate all of the configuration in your struct. It will
// automatically be called when Athens starts
type Validator interface {
	Validate() error
}
