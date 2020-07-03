package stash

import (
	"context"
	"database/sql"
	"fmt"
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

var _postgresAddr string

func postgresAddr(t *testing.T) string {
	addr := os.Getenv("ATHENS_POSTGRES_TCP_ADDR")
	if addr != "" {
		return addr
	}
	out, err := exec.Command("docker-compose", "-p", "athensdev", "port", "postgres", "5432").Output()
	require.NoError(t, err)
	_postgresAddr = strings.TrimSpace(string(out))
	return _postgresAddr
}

// TestWithPostgresLock will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestWithPostgresLock(t *testing.T) {
	createPostgresTestDB(t)
	postgresConfig := getPostgresTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithPostgresLock(postgresConfig, storage.WithChecker(strg))
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

func getPostgresTestConfig(t *testing.T) *config.Postgres {
	t.Helper()

	c, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := c.Index.Postgres
	addr := strings.Split(postgresAddr(t), ":")
	if len(addr) != 2 {
		t.Log("invalid postgres addr", postgresAddr(t))
		t.FailNow()
	}
	cfg.Host = addr[0]

	cfg.Port, err = strconv.Atoi(addr[1])
	require.NoError(t, err)
	cfg.Password = "postgres"

	if os.Getenv("ATHENS_POSTGRES__USER") != "" {
		cfg.User = os.Getenv("ATHENS_POSTGRES__USER")
	}
	if os.Getenv("ATHENS_POSTGRES_PASSWORD") != "" {
		cfg.Password = os.Getenv("ATHENS_POSTGRES_PASSWORD")
	}

	return cfg
}

func createPostgresTestDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if os.Getenv("ATHENS_CREATE_POSTGRES_TEST_DB") == "" {
		return
	}
	cfg := getPostgresTestConfig(t)
	args := []string{}
	args = append(args, "host="+cfg.Host)
	args = append(args, "port=", strconv.Itoa(cfg.Port))
	args = append(args, "user=", cfg.User)
	args = append(args, "password="+cfg.Password)
	for k, v := range cfg.Params {
		args = append(args, k+"="+v)
	}
	dsn := strings.Join(args, " ")
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.PingContext(ctx))
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", cfg.Database))
	require.NoError(t, err)
	exists := rows.Next()
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	if !exists {
		_, err = db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, cfg.Database))
		require.NoError(t, err)
	}
	require.NoError(t, db.Close())
}
