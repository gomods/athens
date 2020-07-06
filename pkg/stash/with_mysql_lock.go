package stash

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/willabides/mysqllocker"
)

const mysqlLockName = "athens_stash_lock"

// WithMysqlLock returns a distributed singleflight using a mysql advisory lock.
func WithMysqlLock(cfg *config.MySQL, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithMysqlLock"
	db, err := sql.Open("mysql", mysqlDSN(cfg))
	if err != nil {
		return nil, errors.E(op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.E(op, err)
	}
	lckr := &mysqlLock{db: db}
	return withLocker(lckr, checker), nil
}

func mysqlDSN(cfg *config.MySQL) string {
	mysqlCfg := mysql.NewConfig()
	mysqlCfg.Net = cfg.Protocol
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Params = cfg.Params
	dsn := mysqlCfg.FormatDSN()
	return dsn
}

type mysqlLock struct {
	db *sql.DB
}

func (l *mysqlLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	return mysqllocker.Lock(ctx, l.db, name,
		mysqllocker.WithTimeout(defaultGetLockTimeout),
		mysqllocker.WithPingInterval(defaultPingInterval),
	)
}
