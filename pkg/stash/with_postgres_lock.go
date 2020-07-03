package stash

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	// register the driver with database/sql
	_ "github.com/lib/pq"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/willabides/pglocker"
)

const postgresLockName = "athens_stash_lock"

// WithPostgresLock returns a distributed singleflight using a postgres advisory lock.
func WithPostgresLock(cfg *config.Postgres, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithPostgresLock"
	db, err := sql.Open("postgres", getPostgresSource(cfg))
	if err != nil {
		return nil, errors.E(op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.E(op, err)
	}
	return func(s Stasher) Stasher {
		return &postgresLock{
			db:      db,
			checker: checker,
			stasher: s,
		}
	}, nil
}

type postgresLock struct {
	db      *sql.DB
	stasher Stasher
	checker storage.Checker
}

func (s *postgresLock) Stash(ctx context.Context, mod, ver string) (string, error) {
	const op errors.Op = "postgresLock.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	lockErrs, err := pglocker.Lock(ctx, s.db, postgresLockName, pglocker.WithTimeout(5*time.Minute))
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

func getPostgresSource(cfg *config.Postgres) string {
	args := []string{}
	args = append(args, "host="+cfg.Host)
	args = append(args, "port=", strconv.Itoa(cfg.Port))
	args = append(args, "user=", cfg.User)
	args = append(args, "dbname=", cfg.Database)
	args = append(args, "password="+cfg.Password)
	for k, v := range cfg.Params {
		args = append(args, k+"="+v)
	}
	return strings.Join(args, " ")
}
