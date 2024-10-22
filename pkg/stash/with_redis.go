package stash

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// RedisLogger mirrors github.com/go-redis/redis/v8/internal.Logging.
type RedisLogger interface {
	Printf(ctx context.Context, format string, v ...any)
}

var errPasswordsDoNotMatch = goerrors.New("a redis url was parsed that contained a password but the configuration also defined a specific redis password, please ensure these values match or use only one of them")

// getRedisClientOptions takes an endpoint and password and returns *redis.Options to use
// with the redis client. endpoint may be a redis url or host:port combination. If a redis
// url is used and a password is also used this function checks to make sure the parsed redis
// url has produced the same password. Preferably, one should use EITHER a redis url or a host:port
// combination w/password but not both. More information on the redis url structure can be found
// here: https://github.com/redis/redis-specifications/blob/master/uri/redis.txt
func getRedisClientOptions(endpoint, password string) (*redis.Options, error) {
	// Try parsing the endpoint as a redis url first. The redis library does not define
	// a specific error when parsing the url so we fall back on the old config here
	// which passed in arguments.
	options, err := redis.ParseURL(endpoint)
	if err != nil {
		return &redis.Options{ //nolint:nilerr // We are specifically falling back here and ignoring the error on purpose.
			Network:  "tcp",
			Addr:     endpoint,
			Password: password,
		}, nil
	}

	// Ensure the password is either empty or that it matches the password
	// parsed from the url into redis.Options. This ensures that if the
	// config supplies the password but a redis url doesn't the behavior
	// is clear vs. failing later on at the time of the first connection
	// with an 'invalid password' like error.
	if password != "" && options.Password != password {
		return nil, errPasswordsDoNotMatch
	}

	return options, nil
}

// WithRedisLock returns a distributed singleflight
// using a redis cluster. If it cannot connect, it will return an error.
func WithRedisLock(l RedisLogger, endpoint, password string, checker storage.Checker, lockConfig *config.RedisLockConfig) (Wrapper, error) {
	redis.SetLogger(l)

	const op errors.Op = "stash.WithRedisLock"

	options, err := getRedisClientOptions(endpoint, password)
	if err != nil {
		return nil, errors.E(op, err)
	}

	client := redis.NewClient(options)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, errors.E(op, err)
	}

	lockOptions, err := lockOptionsFromConfig(lockConfig)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return func(s Stasher) Stasher {
		return &redisLock{client, s, checker, lockOptions}
	}, nil
}

func lockOptionsFromConfig(lockConfig *config.RedisLockConfig) (redisLockOptions, error) {
	if lockConfig.TTL <= 0 || lockConfig.Timeout <= 0 || lockConfig.MaxRetries <= 0 {
		return redisLockOptions{}, goerrors.New("invalid lock options")
	}
	return redisLockOptions{
		ttl:        time.Duration(lockConfig.TTL) * time.Second,
		timeout:    time.Duration(lockConfig.Timeout) * time.Second,
		maxRetries: lockConfig.MaxRetries,
	}, nil
}

type redisLockOptions struct {
	ttl        time.Duration
	timeout    time.Duration
	maxRetries int
}

type redisLock struct {
	client  *redis.Client
	stasher Stasher
	checker storage.Checker
	options redisLockOptions
}

func (s *redisLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "redis.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	mv := config.FmtModVer(mod, ver)
	lockCtx, cancel := context.WithTimeout(ctx, s.options.timeout)
	defer cancel()

	// Obtain a new lock using lock options
	lock, err := redislock.Obtain(lockCtx, s.client, mv, s.options.ttl, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(time.Second), s.options.maxRetries),
	})
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "redis.Release"
		lockErr := lock.Release(ctx)
		if err == nil && lockErr != nil {
			err = errors.E(op, lockErr)
		}
	}()
	ok, err := s.checker.Exists(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	if ok {
		return ver, nil
	}
	newVer, err = s.stasher.Stash(ctx, mod, ver)
	if err != nil {
		return ver, errors.E(op, err)
	}
	return newVer, nil
}
