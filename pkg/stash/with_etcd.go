package stash

import (
	"context"
	stderrors "errors"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	etcdSessionTTL    = 60
	etcdGrantTimeout  = 5 * time.Second
	etcdUnlockTimeout = 5 * time.Second
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
		return &etcd{
			client:        c,
			stasher:       s,
			checker:       checker,
			locks:         make(map[string]chan struct{}),
			unlockTimeout: etcdUnlockTimeout,
		}
	}, nil
}

type etcd struct {
	client        *clientv3.Client
	stasher       Stasher
	checker       storage.Checker
	unlockTimeout time.Duration

	mu       sync.Mutex
	session  *concurrency.Session
	creating chan struct{}
	locks    map[string]chan struct{}
}

func (s *etcd) acquireLocal(ctx context.Context, key string) error {
	for {
		s.mu.Lock()
		ch, busy := s.locks[key]
		if !busy {
			s.locks[key] = make(chan struct{})
			s.mu.Unlock()
			return nil
		}
		s.mu.Unlock()

		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *etcd) releaseLocal(key string) {
	s.mu.Lock()
	ch := s.locks[key]
	delete(s.locks, key)
	s.mu.Unlock()
	if ch != nil {
		close(ch)
	}
}

func (s *etcd) getSession(ctx context.Context) (*concurrency.Session, error) {
	for {
		s.mu.Lock()
		if s.session != nil {
			select {
			case <-s.session.Done():
				s.session = nil
			default:
				sess := s.session
				s.mu.Unlock()
				return sess, nil
			}
		}
		if s.creating != nil {
			wait := s.creating
			s.mu.Unlock()
			select {
			case <-wait:
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		creating := make(chan struct{})
		s.creating = creating
		s.mu.Unlock()

		sess, err := s.newSession()

		s.mu.Lock()
		if err == nil {
			s.session = sess
		}
		s.creating = nil
		close(creating)
		s.mu.Unlock()
		return sess, err
	}
}

func (s *etcd) newSession() (*concurrency.Session, error) {
	grantCtx, cancel := context.WithTimeout(s.client.Ctx(), etcdGrantTimeout)
	resp, err := s.client.Grant(grantCtx, etcdSessionTTL)
	cancel()
	if err != nil {
		return nil, err
	}
	sess, err := concurrency.NewSession(s.client, concurrency.WithLease(resp.ID))
	if err != nil {
		revokeCtx, revokeCancel := context.WithTimeout(s.client.Ctx(), etcdGrantTimeout)
		_, _ = s.client.Revoke(revokeCtx, resp.ID)
		revokeCancel()
		return nil, err
	}
	return sess, nil
}

func (s *etcd) invalidateSession(sess *concurrency.Session) {
	if sess == nil {
		return
	}
	s.mu.Lock()
	if s.session == sess {
		s.session = nil
	}
	s.mu.Unlock()
	sess.Orphan()
}

func isSessionError(err error) bool {
	return stderrors.Is(err, concurrency.ErrSessionExpired) ||
		stderrors.Is(rpctypes.Error(err), rpctypes.ErrLeaseNotFound)
}

func (s *etcd) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "etcd.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	if err := ctx.Err(); err != nil {
		return "", errors.E(op, err)
	}

	key := config.FmtModVer(mod, ver)
	// Concurrency Mutex keys locks by prefix+leaseID, so two goroutines on the
	// same session would re-enter the same lock without local serialization.
	if err := s.acquireLocal(ctx, key); err != nil {
		return "", errors.E(op, err)
	}
	defer s.releaseLocal(key)

	var (
		sess *concurrency.Session
		m    *concurrency.Mutex
	)
	for attempt := range 2 {
		sess, err = s.getSession(ctx)
		if err != nil {
			return "", errors.E(op, err)
		}
		m = concurrency.NewMutex(sess, key)
		err = m.Lock(ctx)
		if err == nil {
			break
		}
		if !isSessionError(err) || attempt == 1 {
			if isSessionError(err) {
				s.invalidateSession(sess)
			}
			return "", errors.E(op, err)
		}
		s.invalidateSession(sess)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", errors.E(op, ctxErr)
		}
	}
	defer func() {
		// Detach from caller cancellation so a canceled request still releases the distributed lock.
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.unlockTimeout)
		defer cancel()
		if unlockErr := m.Unlock(unlockCtx); unlockErr != nil {
			s.invalidateSession(sess)
		}
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
