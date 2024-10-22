package stash

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"golang.org/x/sync/errgroup"
)

// testingRedisLogger implements pkg/stash.RedisLogger.
type testingRedisLogger struct {
	t *testing.T
}

func (l *testingRedisLogger) Printf(ctx context.Context, format string, v ...any) {
	l.t.Logf(format, v...)
}

// WithRedisLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithRedisLock(t *testing.T) {
	endpoint := os.Getenv("REDIS_TEST_ENDPOINT")
	password := os.Getenv("ATHENS_REDIS_PASSWORD")
	if len(endpoint) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	l := &testingRedisLogger{t: t}
	wrapper, err := WithRedisLock(l, endpoint, password, storage.WithChecker(strg), config.DefaultRedisLockConfig())
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

// Verify with WithRedisLock working with password protected redis
// Same logic as the TestWithRedisLock test.
func TestWithRedisLockWithPassword(t *testing.T) {
	endpoint := os.Getenv("PROTECTED_REDIS_TEST_ENDPOINT")
	password := os.Getenv("ATHENS_PROTECTED_REDIS_PASSWORD")
	if len(endpoint) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	l := &testingRedisLogger{t: t}
	wrapper, err := WithRedisLock(l, endpoint, password, storage.WithChecker(strg), config.DefaultRedisLockConfig())
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

// Verify the WithRedisLock fails with the correct error when trying
// to connect with the wrong password.
func TestWithRedisLockWithWrongPassword(t *testing.T) {
	endpoint := os.Getenv("PROTECTED_REDIS_TEST_ENDPOINT")
	password := ""
	if len(endpoint) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	l := &testingRedisLogger{t: t}
	_, err = WithRedisLock(l, endpoint, password, storage.WithChecker(strg), config.DefaultRedisLockConfig())
	if err == nil {
		t.Fatal("Expected Connection Error")
	}

	if !strings.Contains(err.Error(), "NOAUTH Authentication required.") {
		t.Fatalf("Wrong error was thrown %q\n", err.Error())
	}
}

type getRedisClientOptionsFacet struct {
	endpoint string
	password string
	options  *redis.Options
	err      error
}

func Test_getRedisClientOptions(t *testing.T) {
	facets := []*getRedisClientOptionsFacet{
		{
			endpoint: "127.0.0.1:6379",
			options: &redis.Options{
				Addr: "127.0.0.1:6379",
			},
		},
		{
			endpoint: "127.0.0.1:6379",
			password: "1234",
			options: &redis.Options{
				Addr:     "127.0.0.1:6379",
				Password: "1234",
			},
		},
		{
			endpoint: "rediss://username:password@127.0.0.1:6379",
			password: "1234", // Ignored because password was parsed
			err:      errors.E("stash.WithRedisLock", errPasswordsDoNotMatch),
		},
		{
			endpoint: "rediss://username:password@127.0.0.1:6379",
			password: "1234", // Ignored because password was parsed
			err:      errors.E("stash.WithRedisLock", errPasswordsDoNotMatch),
		},
	}

	for i, facet := range facets {
		options, err := getRedisClientOptions(facet.endpoint, facet.password)
		if err != nil && facet.err == nil {
			t.Errorf("Facet %d: no error produced", i)
			continue
		}
		if facet.err != nil {
			if err == nil {
				t.Errorf("Facet %d: no error produced", i)
			} else {
				if err.Error() != facet.err.Error() {
					t.Errorf("Facet %d: expected %q, got %q", i, facet.err, err)
				}
			}
		}

		if err != nil {
			continue
		}
		if facet.options.Addr != options.Addr {
			t.Errorf("Facet %d: Expected Addr %q, got %q", i, facet.options.Addr, options.Addr)
		}
		if facet.options.Username != options.Username {
			t.Errorf("Facet %d: Expected Username %q, got %q", i, facet.options.Username, options.Username)
		}
		if facet.options.Password != options.Password {
			t.Errorf("Facet %d: Expected Password %q, got %q", i, facet.options.Password, options.Password)
		}

	}
}

// mockRedisStasher is like mockStasher
// but leverages in memory storage
// so that redis can determine
// whether to call the underlying stasher or not.
type mockRedisStasher struct {
	strg storage.Backend
	mu   sync.Mutex
	num  int
}

func (ms *mockRedisStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	time.Sleep(time.Millisecond * 100) // allow for second requests to come in.
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.num == 0 {
		err := ms.strg.Save(
			ctx,
			mod,
			ver,
			[]byte("mod file"),
			strings.NewReader("zip file"),
			[]byte("info file"),
		)
		if err != nil {
			return "", err
		}
		ms.num++
		return "", nil
	}
	return "", fmt.Errorf("second time error")
}
