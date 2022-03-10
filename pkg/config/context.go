package config

import (
	"context"
	"time"
)

// noCancel context will ignore any up-the-chain cancellations of the context
//
// We need this in "protocol.processDownload" when using "async" or "async_redirect" download mode
// to avoid the HTTP servers cancellation of the root context due to the browser being redirected
// (in the case of async_redirect) or sent an error (in the case of "async"), closing the connection.
//
// Contexts wrapping this one will behave as normal, including cancellations, but it will sever the life-cycle
// of the root context life-cycle from the async work needed to be done regardless of the client listening or not
type noCancel struct {
	context.Context
}

func (c noCancel) Deadline() (time.Time, bool)       { return c.Context.Deadline() }
func (c noCancel) Done() <-chan struct{}             { return nil }
func (c noCancel) Err() error                        { return c.Context.Err() }
func (c noCancel) Value(key interface{}) interface{} { return c.Context.Value(key) }

// WithoutCancel returns a context that is never canceled.
func ContextWithoutCancel(ctx context.Context) context.Context {
	return noCancel{ctx}
}
