package config

import "time"

// TimeoutConf is a common struct for anything with a timeout
type TimeoutConf struct {
	Timeout int `validate:"required"`
}

// TimeoutDuration returns the timeout as time.duration
func (t *TimeoutConf) TimeoutDuration() time.Duration {
	return GetTimeoutDuration(t.Timeout)
}

// GetTimeoutDuration returns the timeout as time.duration
func GetTimeoutDuration(timeout int) time.Duration {
	return time.Second * time.Duration(timeout)
}
