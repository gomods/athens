package storage

import (
	"context"
	"time"
)

type ListCacher interface {
	ExpiresIn(ctx context.Context, module string) (time.Duration, error)
}
