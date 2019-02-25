package config

// SingleFlight holds the various
// backend configurations for a distributed
// lock or single flight mechanism.
type SingleFlight struct {
	Etcd *Etcd
}

// Etcd holds client side configuration
// that helps Athens connect to the
// Etcd backends.
type Etcd struct {
	Endpoints string `envconfig:"ATHENS_ETCD_ENDPOINTS"`
}
