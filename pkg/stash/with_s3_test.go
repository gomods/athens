package stash

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	s3storage "github.com/gomods/athens/pkg/storage/s3"
	"golang.org/x/sync/errgroup"
)

// TestWithS3Lock ensures that 5 concurrent requests all get the first request's
// response: only the first call to the underlying stasher succeeds, so if the lock
// let more than one through the errgroup would surface the "second time error".
func TestWithS3Lock(t *testing.T) {
	s3Conf, client := getS3LockTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockRedisStasher{strg: strg}
	wrapper, err := WithS3Lock(s3Conf, config.DefaultS3Config(), storage.WithChecker(strg))
	if err != nil {
		t.Fatal(err)
	}
	s := wrapper(ms)

	const mod, ver = "mod", "ver"
	var eg errgroup.Group
	for range 5 {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(t.Context(), time.Second*10)
			defer cancel()
			_, err := s.Stash(ctx, mod, ver)
			return err
		})
	}
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	assertNoLock(t, client, s3Conf.Bucket, mod, ver)
}

// TestWithS3LockHeld ensures a live lock held by another instance blocks the stash
// instead of letting a second writer through.
func TestWithS3LockHeld(t *testing.T) {
	s3Conf, client := getS3LockTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	const mod, ver = "heldmod", "ver"
	putLock(t, client, s3Conf.Bucket, mod, ver)

	ms := &mockRedisStasher{strg: strg}
	lockConf := &config.S3{TTL: 900, Timeout: 5, MaxRetries: 2}
	wrapper, err := WithS3Lock(s3Conf, lockConf, storage.WithChecker(strg))
	if err != nil {
		t.Fatal(err)
	}
	_, err = wrapper(ms).Stash(t.Context(), mod, ver)
	if err == nil {
		t.Fatal("expected the stash to fail while the lock is held by someone else")
	}
	if ms.num != 0 {
		t.Fatal("the underlying stasher must not run while the lock is held by someone else")
	}
}

// TestWithS3LockStale ensures a lock older than the TTL is reclaimed, so an
// instance that died mid-save does not block the module forever.
func TestWithS3LockStale(t *testing.T) {
	s3Conf, client := getS3LockTestConfig(t)
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	const mod, ver = "stalemod", "ver"
	putLock(t, client, s3Conf.Bucket, mod, ver)
	// LastModified has one-second granularity, so sleep well past a one-second TTL.
	time.Sleep(2500 * time.Millisecond)

	ms := &mockRedisStasher{strg: strg}
	lockConf := &config.S3{TTL: 1, Timeout: 10, MaxRetries: 5}
	wrapper, err := WithS3Lock(s3Conf, lockConf, storage.WithChecker(strg))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrapper(ms).Stash(t.Context(), mod, ver); err != nil {
		t.Fatal(err)
	}
	if ms.num != 1 {
		t.Fatalf("expected the stasher to run once after reclaiming the stale lock, ran %d times", ms.num)
	}
	assertNoLock(t, client, s3Conf.Bucket, mod, ver)
}

func lockKey(mod, ver string) string {
	return s3LockPrefix + config.FmtModVer(mod, ver)
}

func putLock(t *testing.T, client *s3.Client, bucket, mod, ver string) {
	t.Helper()
	key := lockKey(mod, ver)
	_, err := client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader("someone else"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	})
}

func assertNoLock(t *testing.T, client *s3.Client, bucket, mod, ver string) {
	t.Helper()
	_, err := client.HeadObject(t.Context(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(lockKey(mod, ver)),
	})
	var notFound *types.NotFound
	if !errors.AsErr(err, &notFound) {
		t.Fatalf("expected the lock object to be released, got: %v", err)
	}
}

// getS3LockTestConfig points the lock at the MinIO instance from docker-compose, using
// its own bucket so the storage/s3 tests wiping their bucket cannot race this package.
func getS3LockTestConfig(t *testing.T) (*config.S3Config, *s3.Client) {
	t.Helper()
	endpoint := os.Getenv("ATHENS_MINIO_ENDPOINT")
	if endpoint == "" {
		t.SkipNow()
	}
	s3Conf := &config.S3Config{
		Key:            "minio",
		Secret:         "minio123",
		Bucket:         "gomodsawslock",
		Region:         "us-west-1",
		ForcePathStyle: true,
		Endpoint:       endpoint,
	}
	client, err := s3storage.NewClient(s3Conf)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.CreateBucket(t.Context(), &s3.CreateBucketInput{Bucket: aws.String(s3Conf.Bucket)})
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists
		if !errors.AsErr(err, &owned) && !errors.AsErr(err, &exists) {
			t.Fatal(err)
		}
	}
	return s3Conf, client
}
