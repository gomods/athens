package stash

import (
	"context"
	"sync"
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

	mu      sync.Mutex
	session *concurrency.Session
}

func (s *etcd) getSession() (*concurrency.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.session != nil {
		select {
		case <-s.session.Done():
			s.session = nil
		default:
			return s.session, nil
		}
	}

	sess, err := concurrency.NewSession(s.client)
	if err != nil {
		return nil, err
	}
	s.session = sess
	return sess, nil
}

func (s *etcd) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "etcd.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	if err := ctx.Err(); err != nil {
		return "", errors.E(op, err)
	}

	sess, err := s.getSession()
	if err != nil {
		return "", errors.E(op, err)
	}

	m := concurrency.NewMutex(sess, config.FmtModVer(mod, ver))
	if err := m.Lock(ctx); err != nil {
		return "", errors.E(op, err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = m.Unlock(unlockCtx)
	}()

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
