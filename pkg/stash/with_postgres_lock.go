package stash

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/willabides/pglocker"

	// register the driver with database/sql
	_ "github.com/lib/pq"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// WithPostgresLock returns a distributed singleflight using a postgres advisory lock.
func WithPostgresLock(cfg *config.Postgres, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithPostgresLock"

	db, err := sql.Open("postgres", postgresDSN(cfg))
	if err != nil {
		return nil, errors.E(op, err)
	}
	err = db.Ping()
	if err != nil {
		return nil, errors.E(op, err)
	}
	lckr := &postgresLock{db: db}
	return withLocker(lckr, checker), nil
}

func postgresDSN(cfg *config.Postgres) string {
	args := []string{}
	args = append(args, "host="+cfg.Host)
	args = append(args, "port=", strconv.Itoa(cfg.Port))
	args = append(args, "user=", cfg.User)
	args = append(args, "password="+cfg.Password)
	for k, v := range cfg.Params {
		args = append(args, k+"="+v)
	}
	return strings.Join(args, " ")
}

type postgresLock struct {
	db *sql.DB
}

func (l *postgresLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	return pglocker.Lock(ctx, l.db, name,
		pglocker.WithTimeout(defaultGetLockTimeout),
		pglocker.WithPingInterval(defaultPingInterval),
	)
}
