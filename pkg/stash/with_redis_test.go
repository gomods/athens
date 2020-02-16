// +build !unit

package stash

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

var useTestContainers = os.Getenv("ATHENS_USE_TEST_CONTAINERS")
var redisEndpoint string
var redisProtectedEndpoint string
var redisPassword string

func TestMain(m *testing.M) {
	redisPassword = os.Getenv("ATHENS_REDIS_PROTECTED_PASSWORD")
	if redisPassword == "" {
		redisPassword = "mydefaultredispassword"
	}

	if useTestContainers != "1" {
		redisEndpoint = os.Getenv("ATHENS_REDIS_ENDPOINT")
		redisProtectedEndpoint = os.Getenv("ATHENS_REDIS_PROTECTED_ENDPOINT")
		redisPassword = os.Getenv("ATHENS_PROTECTED_REDIS_PASSWORD")
		os.Exit(m.Run())
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{"6379/tcp"},
		Env: map[string]string{
			"REDIS_PASSWORD": redisPassword,
		},
		Cmd:        []string{"sh", "-c", "redis-server --requirepass \"$REDIS_PASSWORD\""},
		WaitingFor: wait.ForLog("Ready to accept connections").WithStartupTimeout(time.Minute * 1),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		os.Exit(1)
	}

	ep, err := c.Endpoint(context.Background(), "")
	if err != nil {
		panic(err.Error())
	}

	redisProtectedEndpoint = ep
	defer c.Terminate(ctx)

	req = testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(time.Minute * 1),
	}

	c, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		os.Exit(1)
	}

	ep, err = c.Endpoint(context.Background(), "")
	if err != nil {
		panic(err.Error())
	}

	redisEndpoint = ep
	defer c.Terminate(ctx)

	os.Exit(m.Run())
}

// WithRedisLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithRedisLock(t *testing.T) {
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithRedisLock(redisEndpoint, "", storage.WithChecker(strg))
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
	endpoint := redisProtectedEndpoint
	password := redisPassword
	if len(endpoint) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithRedisLock(endpoint, password, storage.WithChecker(strg))
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
	endpoint := redisProtectedEndpoint
	password := ""
	if len(endpoint) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	_, err = WithRedisLock(endpoint, password, storage.WithChecker(strg))
	if err == nil {
		t.Fatal("Expected Connection Error")
	}

	if !strings.Contains(err.Error(), "NOAUTH Authentication required.") {
		t.Fatalf("Wrong error was thrown %q\n", err.Error())
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
