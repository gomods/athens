package stash

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/google/uuid"
)

// WithAzureBlobLock returns a distributed singleflight
// using a Azure Blob Storage backend. See the config.toml documentation for details.
func WithAzureBlobLock(conf *config.AzureBlobConfig, timeout time.Duration, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithAzureBlobLock"

	accountURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	cred, err := azblob.NewSharedKeyCredential(conf.AccountName, conf.AccountKey)
	if err != nil {
		return nil, errors.E(op, err)
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*accountURL, pipe)

	lckr := &azBlobLock{
		containerURL: serviceURL.NewContainerURL(conf.ContainerName),
	}
	return withLocker(lckr, checker), nil
}

type azBlobLock struct {
	containerURL azblob.ContainerURL
}

func (l *azBlobLock) lock(ctx context.Context, name string) (releaseErrs <-chan error, err error) {
	ttl := defaultPingInterval * 2
	const op errors.Op = "azBlobLock.lock"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	leaseBlobName := "lease/" + name
	leaseBlobURL := l.containerURL.NewBlockBlobURL(leaseBlobName)
	acquireCtx, acquireCancel := context.WithTimeout(ctx, defaultGetLockTimeout)
	defer acquireCancel()
	leaseID, err := azAcquireLease(acquireCtx, leaseBlobURL, ttl)
	if err != nil {
		return nil, errors.E(op, err)
	}
	holder := &lockHolder{
		pingInterval: defaultPingInterval,
		ttl:          ttl,
		refresh: func(refreshCtx context.Context) error {
			_, refreshErr := leaseBlobURL.RenewLease(refreshCtx, leaseID, azblob.ModifiedAccessConditions{})
			return refreshErr
		},
		release: func() error {
			_, releaseErr := leaseBlobURL.ReleaseLease(context.Background(), leaseID, azblob.ModifiedAccessConditions{})
			return releaseErr
		},
	}
	errs := make(chan error, 1)
	go holder.holdAndRelease(ctx, errs)
	return errs, nil
}

func azAcquireLease(ctx context.Context, blobURL azblob.BlockBlobURL, ttl time.Duration) (string, error) {
	const op errors.Op = "azAcquireLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	tctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// first we need to create a blob which can be then leased
	_, err := blobURL.Upload(tctx, bytes.NewReader([]byte{1}), azblob.BlobHTTPHeaders{}, nil, azblob.BlobAccessConditions{})
	if err != nil {
		// if the blob is already leased we will get http.StatusPreconditionFailed while writing to that blob
		stgErr, ok := err.(azblob.StorageError)
		if !ok || stgErr.Response().StatusCode != http.StatusPreconditionFailed {
			return "", errors.E(op, err)
		}
	}

	leaseID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.E(op, err)
	}
	ttlSeconds := int32(ttl / time.Second)
	for {
		// acquire lease for 20 sec
		res, err := blobURL.AcquireLease(tctx, leaseID.String(), ttlSeconds, azblob.ModifiedAccessConditions{})
		if err != nil {
			// if the blob is already leased we will get http.StatusConflict - wait and try again
			if stgErr, ok := err.(azblob.StorageError); ok && stgErr.Response().StatusCode == http.StatusConflict {
				select {
				case <-time.After(1 * time.Second):
					continue
				case <-tctx.Done():
					return "", tctx.Err()
				}
			}
			return "", errors.E(op, err)
		}
		return res.LeaseID(), nil
	}
}
