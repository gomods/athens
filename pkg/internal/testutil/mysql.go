package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/require"
)

var (
	mysqlTestConfigOnce sync.Once
	mysqlTestConfig     *config.MySQL
)

// MySQLTestConfig returns a *config.MySQL to be used in tests. It creates the mysql database if it doesn't already exist.
func MySQLTestConfig(t *testing.T) *config.MySQL {
	t.Helper()
	mysqlTestConfigOnce.Do(func() {
		mysqlAddr := os.Getenv("ATHENS_MYSQL_TCP_ADDR")
		if mysqlAddr == "" {
			out, err := exec.Command("docker-compose", "-p", "athensdev", "port", "mysql", "3306").Output()
			require.NoError(t, err)
			mysqlAddr = strings.TrimSpace(string(out))
		}
		addr := strings.Split(mysqlAddr, ":")
		if len(addr) != 2 {
			t.Log("invalid mysql addr", mysqlAddr)
			t.FailNow()
		}
		c, err := config.Load("")
		require.NoError(t, err)
		mysqlTestConfig = c.Index.MySQL
		mysqlTestConfig.Host = addr[0]
		mysqlTestConfig.Port, err = strconv.Atoi(addr[1])
		require.NoError(t, err)
		if os.Getenv("ATHENS_MYSQL_USER") != "" {
			mysqlTestConfig.User = os.Getenv("ATHENS_MYSQL_USER")
		}
		if os.Getenv("ATHENS_MYSQL_PASSWORD") != "" {
			mysqlTestConfig.Password = os.Getenv("ATHENS_MYSQL_PASSWORD")
		}
		createMySQLTestDatabase(t, mysqlTestConfig)
	})
	cfg := new(config.MySQL)
	*cfg = *mysqlTestConfig
	cfg.Params = make(map[string]string, len(mysqlTestConfig.Params))
	for k, v := range mysqlTestConfig.Params {
		cfg.Params[k] = v
	}
	return cfg
}

func createMySQLTestDatabase(t *testing.T, cfg *config.MySQL) {
	t.Helper()
	c := mySQLConfigDSN(cfg)
	dbName := c.DBName
	c.DBName = ""
	db, err := sql.Open("mysql", c.FormatDSN())
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName))
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func mySQLConfigDSN(cfg *config.MySQL) *mysql.Config {
	c := mysql.NewConfig()
	c.Net = cfg.Protocol
	c.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	c.User = cfg.User
	c.Passwd = cfg.Password
	c.Params = cfg.Params
	c.DBName = cfg.Database
	return c
}
