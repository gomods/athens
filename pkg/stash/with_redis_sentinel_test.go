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
	sentinelPassword := os.Getenv("REDIS_SENTINEL_TEST_PASSWORD")
	if len(endpoint) == 0 || len(masterName) == 0 || len(sentinelPassword) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	l := &testingRedisLogger{t: t}

	wrapper, err := WithRedisSentinelLock(l, []string{endpoint}, masterName, sentinelPassword, "", "", storage.WithChecker(strg), config.DefaultRedisLockConfig())
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

// TestWithRedisSentinelLockWithRedisPassword verifies WithRedisSentinelLock working with
// password protected redis sentinel and redis master nodes
func TestWithRedisSentinelLockWithRedisPassword(t *testing.T) {
	endpoint := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_ENDPOINT")
	masterName := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_MASTER_NAME")
	sentinelPassword := os.Getenv("REDIS_SENTINEL_TEST_PASSWORD")
	redisPassword := os.Getenv("ATHENS_PROTECTED_REDIS_PASSWORD")
	if len(endpoint) == 0 || len(masterName) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	l := &testingRedisLogger{t: t}
	wrapper, err := WithRedisSentinelLock(l, []string{endpoint}, masterName, sentinelPassword, "", redisPassword, storage.WithChecker(strg), config.DefaultRedisLockConfig())
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

// TestWithRedisSentinelLockWithPassword verifies WithRedisSentinelLock working with
// username & password protected master node under the redis sentinel
func TestWithRedisSentinelLockWithUsernameAndPassword(t *testing.T) {
	endpoint := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_ENDPOINT")
	masterName := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_MASTER_NAME")
	sentinelPassword := os.Getenv("REDIS_SENTINEL_TEST_PASSWORD")
	redisPassword := os.Getenv("ATHENS_PROTECTED_REDIS_PASSWORD")
	redisUsername := os.Getenv("PROTECTED_REDIS_TEST_USERNAME")
	if len(endpoint) == 0 || len(masterName) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	l := &testingRedisLogger{t: t}
	wrapper, err := WithRedisSentinelLock(l, []string{endpoint}, masterName, sentinelPassword, redisUsername, redisPassword, storage.WithChecker(strg), config.DefaultRedisLockConfig())
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

// TestWithRedisSentinelLockWithWrongPassword verifies the WithRedisSentinelLock fails
// with the correct error when trying to connect with wrong passwords
func TestWithRedisSentinelLockWithWrongRedisPassword(t *testing.T) {
	endpoint := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_ENDPOINT")
	masterName := os.Getenv("REDIS_SENTINEL_TEST_PROTECTED_MASTER_NAME")
	sentinelPassword := os.Getenv("REDIS_SENTINEL_TEST_PASSWORD")
	redisPassword := os.Getenv("ATHENS_PROTECTED_REDIS_PASSWORD")
	if len(endpoint) == 0 || len(masterName) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	l := &testingRedisLogger{t: t}

	// Test with wrong sentinel password
	_, err = WithRedisSentinelLock(l, []string{endpoint}, masterName, "wrong-sentinel-password", "", redisPassword, storage.WithChecker(strg), config.DefaultRedisLockConfig())
	if err == nil {
		t.Fatal("Expected Connection Error for wrong sentinel password")
	}

	// Test with wrong redis password
	_, err = WithRedisSentinelLock(l, []string{endpoint}, masterName, sentinelPassword, "", "wrong-redis-password", storage.WithChecker(strg), config.DefaultRedisLockConfig())
	if err == nil {
		t.Fatal("Expected Connection Error for wrong redis password")
	}
}
