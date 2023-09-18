package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// WithEtcd returns a distributed singleflight
// using an etcd cluster. If it cannot connect,
// to any of the endpoints, it will return an error.
func WithEtcd(endpoints []string, checker storage.Checker) (Wrapper, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, errors.E("stash.WithEtcd", err)
	}
	return func(s Stasher) Stasher {
		return &etcd{client: c, stasher: s, checker: checker}
	}, nil
}

type etcd struct {
	client  *clientv3.Client
	stasher Stasher
	checker storage.Checker
}

func (s *etcd) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "etcd.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	sess, err := concurrency.NewSession(s.client)
	if err != nil {
		return "", errors.E(op, err)
	}
	defer sess.Close()

	m := concurrency.NewMutex(sess, config.FmtModVer(mod, ver))
	if err := m.Lock(ctx); err != nil {
		return "", errors.E(op, err)
	}
	defer m.Unlock(ctx)

	ok, err := s.checker.Exists(ctx, mod, ver)
	if err != nil {
		return "", errors.E(op, err)
	}

	if ok {
		return ver, nil
	}

	newVer, err = s.stasher.Stash(ctx, mod, ver)
	if err != nil {
		return "", errors.E(op, err)
	}
	return newVer, nil
}
