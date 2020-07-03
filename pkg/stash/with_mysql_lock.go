package stash

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/willabides/mysqllocker"
)

const mysqlLockName = "athens_stash_lock"

// WithMysqlLock returns a distributed singleflight using a mysql advisory lock.
func WithMysqlLock(cfg *config.MySQL, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithMysqlLock"
	db, err := sql.Open("mysql", getMySQLSource(cfg))
	if err != nil {
		return nil, errors.E(op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return func(s Stasher) Stasher {
		return &mysqlLock{
			db:      db,
			checker: checker,
			stasher: s,
		}
	}, nil
}

type mysqlLock struct {
	db      *sql.DB
	stasher Stasher
	checker storage.Checker
}

func (s *mysqlLock) Stash(ctx context.Context, mod string, ver string) (string, error) {
	const op errors.Op = "mysqlLock.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	lockErrs, err := mysqllocker.Lock(ctx, s.db, mysqlLockName, mysqllocker.WithTimeout(5*time.Minute))
	if err != nil {
		return ver, errors.E(op, err)
	}
	ok, err := s.checker.Exists(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	var newVer string
	if ok {
		newVer = ver
	} else {
		newVer, err = s.stasher.Stash(ctx, mod, ver)
		if err != nil {
			return ver, errors.E(op, err)
		}
	}
	cancel()
	err = <-lockErrs
	if err != nil {
		return newVer, errors.E(op, fmt.Errorf("could not release lock: %w", err))
	}
	return newVer, nil
}

func getMySQLSource(cfg *config.MySQL) string {
	c := mysql.NewConfig()
	c.Net = cfg.Protocol
	c.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	c.User = cfg.User
	c.Passwd = cfg.Password
	c.DBName = cfg.Database
	c.Params = cfg.Params
	return c.FormatDSN()
}
