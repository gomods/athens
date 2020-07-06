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

	mysqlCfg := mysql.NewConfig()
	mysqlCfg.Net = cfg.Protocol
	mysqlCfg.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mysqlCfg.User = cfg.User
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Params = cfg.Params

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
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

type mysqlLock struct {
	db *sql.DB
}

func (l *mysqlLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	return mysqllocker.Lock(ctx, l.db, name,
		mysqllocker.WithTimeout(defaultGetLockTimeout),
		mysqllocker.WithPingInterval(defaultPingInterval),
	)
}
