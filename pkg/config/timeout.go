package config

import "time"

// TimeoutDuration returns the given timeouts(in seconds) as time.Duration
func TimeoutDuration(t int) time.Duration {
	return time.Duration(t) * time.Second
}
