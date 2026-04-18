package config

import "time"

// TimeoutConfig is a common struct for anything with a timeout.
type TimeoutConfig struct {
	Timeout         int `envconfig:"ATHENS_TIMEOUT"          validate:"required"`
	ShutdownTimeout int `envconfig:"ATHENS_SHUTDOWN_TIMEOUT" validate:"min=0"`
	StashTimeout    int `envconfig:"ATHENS_STASH_TIMEOUT"`
}

// TimeoutDuration returns the timeout as time.Duration.
func (t *TimeoutConfig) TimeoutDuration() time.Duration {
	return timeoutAsSeconds(t.Timeout)
}

// ShutdownTimeoutDuration return the shutdown timeout as a time.Duration.
func (t *TimeoutConfig) ShutdownTimeoutDuration() time.Duration {
	return timeoutAsSeconds(t.ShutdownTimeout)
}

// StashTimeoutDuration returns the stash timeout as time.Duration.
func (c *TimeoutConfig) StashTimeoutDuration() time.Duration {
	return timeoutAsSeconds(c.StashTimeout)
}

func timeoutAsSeconds(timeout int) time.Duration {
	return time.Second * time.Duration(timeout)
}
