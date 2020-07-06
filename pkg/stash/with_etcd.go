package stash

import (
	"context"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"golang.org/x/sync/errgroup"
)

// WithEtcd returns a distributed singleflight
// using an etcd cluster. If it cannot connect,
// to any of the endpoints, it will return an error.
func WithEtcd(endpoints []string, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithEtcd"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	var eg errgroup.Group
	for _, ep := range endpoints {
		eg.Go(func() error {
			_, err := c.Status(ctx, ep)
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, errors.E(op, err)
	}
	lckr := &etcdLock{client: c}
	return withLocker(lckr, checker), nil
}

type etcdLock struct {
	client *clientv3.Client
}

func (l *etcdLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	const op errors.Op = "etcdLock.lock"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	ttl := defaultPingInterval
	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(int(ttl/time.Second)))
	if err != nil {
		return nil, errors.E(op, err)
	}
	mu := concurrency.NewMutex(session, name)
	err = mu.Lock(ctx)
	if err != nil {
		return nil, err
	}
	holder := &lockHolder{
		ttl: ttl,
		release: func() error {
			_ = mu.Unlock(ctx)  // don't care about this error because the lock will expire when the session is closed below
			_ = session.Close() // don't care about this error because close blocks long enough for the lock to expire when there is an error
			return nil
		},
		refresh: func(refreshCtx context.Context) error {
			// etcd does its own keepalive, so there is no need for us to do one too
			return nil
		},
	}
	errs := make(chan error, 1)
	go holder.holdAndRelease(ctx, errs)
	return errs, nil
}
