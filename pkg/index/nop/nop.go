package nop

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/index"
)

// New returns a no-op Indexer
func New() index.Indexer {
	return indexer{}
}

type indexer struct{}

func (indexer) Index(ctx context.Context, mod, ver string) error {
	return nil
}
func (indexer) Lines(ctx context.Context, since time.Time, limit int) ([]*index.Line, error) {
	return []*index.Line{}, nil
}
