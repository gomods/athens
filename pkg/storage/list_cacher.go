package storage

import (
	"context"
	"time"
)

type ListCacher interface {
	Reset(context.Context, string) error
	ExpiresIn(context.Context, string) (time.Duration, error)
}
