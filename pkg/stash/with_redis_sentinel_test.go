package stash

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"golang.org/x/sync/errgroup"
)

// WithRedisLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithRedisSentinelLock(t *testing.T) {
	endpoint := os.Getenv("REDIS_SENTINEL_TEST_ENDPOINT")
	masterName := os.Getenv("REDIS_SENTINEL_TEST_MASTER_NAME")
	password := os.Getenv("REDIS_SENTINEL_TEST_PASSWORD")
	if len(endpoint) == 0 || len(masterName) == 0 || len(password) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}

	wrapper, err := WithRedisSentinelLock([]string{endpoint}, masterName, password, storage.WithChecker(strg), config.DefaultRedisLockConfig())
	if err != nil {
		t.Fatal(err)
	}
	s := wrapper(ms)

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			_, err := s.Stash(ctx, "mod", "ver")
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}
