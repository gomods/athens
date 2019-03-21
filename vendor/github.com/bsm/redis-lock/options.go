package lock

import "time"

// Options describe the options for the lock
type Options struct {
	// The maximum duration to lock a key for
	// Default: 5s
	LockTimeout time.Duration

	// The number of time the acquisition of a lock will be retried.
	// Default: 0 = do not retry
	RetryCount int

	// RetryDelay is the amount of time to wait between retries.
	// Default: 100ms
	RetryDelay time.Duration
}

func (o *Options) normalize() *Options {
	if o.LockTimeout < 1 {
		o.LockTimeout = 5 * time.Second
	}
	if o.RetryCount < 0 {
		o.RetryCount = 0
	}
	if o.RetryDelay < 1 {
		o.RetryDelay = 100 * time.Millisecond
	}
	return o
}
