package index

import (
	"context"
	"time"
)

// Line represents a module@version line
// with its metadata such as creation time.
type Line struct {
	Module, Version string
	Timestamp       time.Time
}

// Indexer is an interface that can process new module@versions
// and also retrieve 'limit' module@versions that were indexed after 'since'
type Indexer interface {
	// Index stores the module@version into the index backend.
	// Implementer must create the Timestamp at the time and set it
	// to the time this method is call.
	Index(ctx context.Context, mod, ver string) error

	// Lines returns the module@version lines given the time and limit
	// constraints
	Lines(ctx context.Context, since time.Time, limit int) ([]*Line, error)
}
