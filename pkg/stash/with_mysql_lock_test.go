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

	"github.com/go-sql-driver/mysql"
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
	createMysqlTestDB(t)
	mysqlConfig := getMysqlTestConfig(t)
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

var _mysqlAddr string

func mysqlAddr(t *testing.T) string {
	t.Helper()
	addr := os.Getenv("ATHENS_MYSQL_TCP_ADDR")
	if addr != "" {
		return addr
	}
	out, err := exec.Command("docker-compose", "-p", "athensdev", "port", "mysql", "3306").Output()
	require.NoError(t, err)
	_mysqlAddr = strings.TrimSpace(string(out))
	return _mysqlAddr
}

func getMysqlTestConfig(t *testing.T) *config.MySQL {
	t.Helper()
	c, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := c.Index.MySQL
	addr := strings.Split(mysqlAddr(t), ":")
	if len(addr) != 2 {
		t.Log("invalid mysql addr", mysqlAddr(t))
		t.FailNow()
	}
	cfg.Host = addr[0]

	cfg.Port, err = strconv.Atoi(addr[1])
	require.NoError(t, err)

	if os.Getenv("ATHENS_MYSQL_USER") != "" {
		cfg.User = os.Getenv("ATHENS_MYSQL_USER")
	}
	if os.Getenv("ATHENS_MYSQL_PASSWORD") != "" {
		cfg.Password = os.Getenv("ATHENS_MYSQL_PASSWORD")
	}

	return cfg
}

func createMysqlTestDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if os.Getenv("ATHENS_CREATE_MYSQL_TEST_DB") == "" {
		return
	}
	cfg := getMysqlTestConfig(t)
	c := mysql.NewConfig()
	c.Net = cfg.Protocol
	c.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	c.User = cfg.User
	c.Passwd = cfg.Password
	c.Params = cfg.Params
	db, err := sql.Open("mysql", c.FormatDSN())
	require.NoError(t, err)
	require.NoError(t, db.PingContext(ctx))
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", cfg.Database))
	require.NoError(t, err)
	require.NoError(t, db.Close())
}
