package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/require"
)

var (
	postgresTestConfigOnce sync.Once
	postgresTestConfig     *config.Postgres
)

// PostgresTestConfig returns a *config.Postgres to be used in tests. It creates the postgres database if it doesn't already exist.
func PostgresTestConfig(t *testing.T) *config.Postgres {
	t.Helper()
	postgresTestConfigOnce.Do(func() {
		pgAddr := os.Getenv("ATHENS_POSTGRES_TCP_ADDR")
		if pgAddr == "" {
			out, err := exec.Command("docker-compose", "-p", "athensdev", "port", "postgres", "5432").Output()
			require.NoError(t, err)
			pgAddr = strings.TrimSpace(string(out))
		}
		addr := strings.Split(pgAddr, ":")
		if len(addr) != 2 {
			t.Log("invalid postgres addr", pgAddr)
			t.FailNow()
		}
		c, err := config.Load("")
		require.NoError(t, err)
		postgresTestConfig = c.Index.Postgres
		postgresTestConfig.Host = addr[0]
		postgresTestConfig.Port, err = strconv.Atoi(addr[1])
		postgresTestConfig.Password = "postgres"
		require.NoError(t, err)
		if os.Getenv("ATHENS_POSTGRES_USER") != "" {
			postgresTestConfig.User = os.Getenv("ATHENS_POSTGRES_USER")
		}
		if os.Getenv("ATHENS_POSTGRES_PASSWORD") != "" {
			postgresTestConfig.Password = os.Getenv("ATHENS_POSTGRES_PASSWORD")
		}
		createPostgresTestDatabase(t, postgresTestConfig)
	})
	cfg := new(config.Postgres)
	*cfg = *postgresTestConfig
	cfg.Params = make(map[string]string, len(postgresTestConfig.Params))
	for k, v := range postgresTestConfig.Params {
		cfg.Params[k] = v
	}
	return cfg
}

func createPostgresTestDatabase(t *testing.T, cfg *config.Postgres) {
	t.Helper()
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
	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock(97867)")
	require.NoError(t, err)
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", cfg.Database))
	require.NoError(t, err)
	exists := rows.Next()
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	if !exists {
		_, err = conn.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, cfg.Database))
		require.NoError(t, err)
	}
	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_unlock(97867)")
	require.NoError(t, err)
	require.NoError(t, conn.Close())
	require.NoError(t, db.Close())
}
