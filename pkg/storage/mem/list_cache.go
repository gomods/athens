package mem

import (
	"context"
	"fmt"
	"time"
)

func (s *stg) ExpiresIn(ctx context.Context, mod string) (time.Duration, error) {
	s.listCacheMut.RLock()
	defer s.listCacheMut.RUnlock()
	expireTime, ok := s.listCache[mod]
	// if the module isn't even in the list cache, then bail
	if !ok {
		return time.Duration(0), fmt.Errorf("No cache entry for module %s", mod)
	}
	dur := expireTime.Sub(time.Now())
	return dur, nil
}

func (s *stg) Reset(ctx context.Context, mod string) error {
	s.listCacheMut.Lock()
	defer s.listCacheMut.Unlock()
	s.listCache[mod] = time.Now().Add(s.ttl)
	return nil
}
