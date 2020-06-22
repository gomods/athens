package mem

import (
	"context"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/index"
)

// New returns a new in-memory indexer
func New() index.Indexer {
	return &indexer{}
}

type indexer struct {
	mu    sync.RWMutex
	lines []*index.Line
}

func (i *indexer) Index(ctx context.Context, mod, ver string) error {
	const op errors.Op = "mem.Index"
	i.mu.Lock()
	i.lines = append(i.lines, &index.Line{
		Module:    mod,
		Version:   ver,
		Timestamp: time.Now(),
	})
	i.mu.Unlock()
	return nil
}

func (i *indexer) Lines(ctx context.Context, since time.Time, limit int) ([]*index.Line, error) {
	const op errors.Op = "mem.Lines"
	lines := []*index.Line{}
	var count int
	i.mu.RLock()
	defer i.mu.RUnlock()
	for _, line := range i.lines {
		if count >= limit {
			break
		}
		if since.After(line.Timestamp) {
			continue
		}
		lines = append(lines, line)
		count++
	}
	return lines, nil
}
