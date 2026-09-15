package stash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	s3storage "github.com/gomods/athens/pkg/storage/s3"
	"github.com/google/uuid"
)

const (
	s3LockPrefix        = "lock/"
	s3LockRetryInterval = time.Second
	s3LockOpTimeout     = 30 * time.Second
)

// WithS3Lock returns a distributed singleflight backed by S3 conditional writes.
// Each module version is guarded by an object under lock/ in the storage bucket that is
// created with If-None-Match: *, so only one instance wins the race to save it.
// See the config.toml documentation for details.
func WithS3Lock(s3Conf *config.S3Config, lockConf *config.S3, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithS3Lock"
	if lockConf == nil || lockConf.TTL <= 0 || lockConf.Timeout <= 0 || lockConf.MaxRetries <= 0 {
		return nil, errors.E(op, fmt.Errorf("invalid lock options"))
	}
	client, err := s3storage.NewClient(s3Conf)
	if err != nil {
		return nil, errors.E(op, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), s3LockOpTimeout)
	defer cancel()
	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s3Conf.Bucket)}); err != nil {
		return nil, errors.E(op, err)
	}
	return func(s Stasher) Stasher {
		return &s3Lock{
			client:     client,
			bucket:     s3Conf.Bucket,
			stasher:    s,
			checker:    checker,
			ttl:        time.Duration(lockConf.TTL) * time.Second,
			timeout:    time.Duration(lockConf.Timeout) * time.Second,
			maxRetries: lockConf.MaxRetries,
		}
	}, nil
}

type s3Lock struct {
	client     *s3.Client
	bucket     string
	stasher    Stasher
	checker    storage.Checker
	ttl        time.Duration
	timeout    time.Duration
	maxRetries int
}

func (s *s3Lock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "s3lock.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	key := s3LockPrefix + config.FmtModVer(mod, ver)
	etag, err := s.acquire(ctx, key)
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "s3lock.Release"
		relErr := s.release(ctx, key, etag)
		if err == nil && relErr != nil {
			err = errors.E(op, relErr)
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

// acquire polls until this instance holds the lock object at key, returning its ETag.
func (s *s3Lock) acquire(ctx context.Context, key string) (string, error) {
	const op errors.Op = "s3lock.acquire"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	tctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	for range s.maxRetries {
		etag, err := s.tryAcquire(tctx, key)
		if err != nil {
			return "", errors.E(op, err)
		}
		if etag != "" {
			return etag, nil
		}
		select {
		case <-time.After(s3LockRetryInterval):
		case <-tctx.Done():
			return "", errors.E(op, tctx.Err())
		}
	}
	return "", errors.E(op, fmt.Errorf("lock %s is still held after %d attempts", key, s.maxRetries))
}

// tryAcquire makes a single attempt at the lock. It returns the ETag of the lock object when this
// instance now holds it, or an empty string when another instance holds a lock that is not yet stale.
func (s *s3Lock) tryAcquire(ctx context.Context, key string) (string, error) {
	etag, err := s.put(ctx, key, &s3.PutObjectInput{IfNoneMatch: aws.String("*")})
	if err == nil {
		return etag, nil
	}
	if !isConditionFailed(err) {
		return "", err
	}
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.AsErr(err, &notFound) {
			// The holder released between the write and this read; the next attempt gets a clean shot.
			return "", nil
		}
		return "", err
	}
	if head.LastModified == nil || time.Since(*head.LastModified) <= s.ttl {
		return "", nil
	}
	// A lock older than the TTL was left behind by an instance that died mid-save. Taking it over is
	// conditional on the ETag we just observed so that only one of the waiting instances succeeds.
	etag, err = s.put(ctx, key, &s3.PutObjectInput{IfMatch: head.ETag})
	if err == nil {
		return etag, nil
	}
	if isConditionFailed(err) {
		return "", nil
	}
	return "", err
}

// put writes the lock object with the given conditions and returns the resulting ETag. The body is a
// fresh UUID so every acquisition yields a distinct ETag; otherwise two instances reclaiming the same
// stale lock with identical bodies would both pass their If-Match check.
func (s *s3Lock) put(ctx context.Context, key string, in *s3.PutObjectInput) (string, error) {
	in.Bucket = aws.String(s.bucket)
	in.Key = aws.String(key)
	in.Body = strings.NewReader(uuid.NewString())
	in.ContentType = aws.String("text/plain")
	out, err := s.client.PutObject(ctx, in)
	if err != nil {
		return "", err
	}
	return aws.ToString(out.ETag), nil
}

// release deletes the lock object only while it is still the one this instance created. A lock that
// was reclaimed by another instance after our TTL expired belongs to that instance and stays put.
// Cancellation of the caller's context is dropped so that an aborted request still frees its lock.
func (s *s3Lock) release(ctx context.Context, key, etag string) error {
	const op errors.Op = "s3lock.release"
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s3LockOpTimeout)
	defer cancel()
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(key),
		IfMatch: aws.String(etag),
	})
	if err != nil && !isConditionFailed(err) {
		return errors.E(op, err)
	}
	return nil
}

// isConditionFailed reports whether S3 refused a conditional request: 412 when the precondition did
// not hold, or 409 when a concurrent conditional write on the same key was still in flight.
func isConditionFailed(err error) bool {
	var apiErr smithy.APIError
	if !errors.AsErr(err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "PreconditionFailed", "ConditionalRequestConflict":
		return true
	default:
		return false
	}
}
