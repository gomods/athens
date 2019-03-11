package config

// SingleFlight holds the various
// backend configurations for a distributed
// lock or single flight mechanism.
type SingleFlight struct {
	Etcd  *Etcd
	Redis *Redis
}

// Etcd holds client side configuration
// that helps Athens connect to the
// Etcd backends.
type Etcd struct {
	Endpoints string `envconfig:"ATHENS_ETCD_ENDPOINTS"`
}

// Redis holds the client side configuration
// to connect to redis as a SingleFlight implementation.
type Redis struct {
	Endpoint string `envconfig:"ATHENS_REDIS_ENDPOINT"`
}
