package stash

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// TestWithMysqlLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithMysqlLock(t *testing.T) {
	mysqlConfig := mysqlTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithMysqlLock(mysqlConfig, storage.WithChecker(strg))
	require.NoError(t, err)
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

func mysqlTestConfig(t *testing.T) *config.MySQL {
	t.Helper()
	c, err := config.Load("")
	require.NoError(t, err)
	cfg := c.Index.MySQL
	if os.Getenv("STATIC_PORTS") == "" {
		cfg.Host, cfg.Port = getServicePort(t, "mysql", 3306)
	}
	return cfg
}

func getServicePort(t *testing.T, service string, containerPort int) (host string, hostPort int) {
	t.Helper()
	cPort := strconv.Itoa(containerPort)
	out, err := exec.Command("docker-compose", "-p", "athensdev", "port", service, cPort).Output()
	require.NoError(t, err)
	addr := strings.TrimSpace(string(out))
	parts := strings.Split(addr, ":")
	require.Lenf(t, parts, 2, "invalid address %q", addr)
	hPort, err := strconv.Atoi(parts[1])
	require.NoError(t, err)
	return parts[0], hPort
}
