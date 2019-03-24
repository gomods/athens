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

func WithAzureBlob(conf *config.AzureBlobConfig, timeout time.Duration, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithAzureBlob"

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
	// rules on container names:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#container-names
	//
	// This container must exist
	containerURL := serviceURL.NewContainerURL(conf.ContainerName)

	return func(s Stasher) Stasher {
		return &azureBlob{containerURL, s, checker}
	}, nil
}

type azureBlob struct {
	containerURL azblob.ContainerURL
	stasher      Stasher
	checker      storage.Checker
}

type stashRes struct {
	v   string
	err error
}

func (s *azureBlob) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "azureblob.Stash"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	leaseBlobName := "lease/" + config.FmtModVer(mod, ver)
	leaseBlobURL := s.containerURL.NewBlockBlobURL(leaseBlobName)

	leaseID, err := s.AcquireLease(ctx, leaseBlobURL)
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "azureblob.Unlock"
		relErr := s.ReleaseLease(ctx, leaseBlobURL, leaseID)
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
	stash := func() <-chan stashRes {
		sChan := make(chan stashRes)
		go func() {
			v, err := s.stasher.Stash(ctx, mod, ver)
			sChan <- stashRes{v, err}
		}()
		return sChan
	}

	for {
		select {
		case sr := <-stash():
			if sr.err != nil {
				err = errors.E(op, sr.err)
				return ver, err
			}
			newVer = sr.v
			return newVer, nil
		case <-time.After(10 * time.Second):
			err := s.RenewLease(ctx, leaseBlobURL, leaseID)
			if err != nil {
				return ver, errors.E(op, err)
			}
		}
	}
}

func (s *azureBlob) ReleaseLease(ctx context.Context, blobURL azblob.BlockBlobURL, leaseID string) error {
	const op errors.Op = "azureblob.ReleaseLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := blobURL.ReleaseLease(ctx, leaseID, azblob.ModifiedAccessConditions{})
	return err
}

func (s *azureBlob) RenewLease(ctx context.Context, blobURL azblob.BlockBlobURL, leaseID string) error {
	const op errors.Op = "azureblob.RenewLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := blobURL.RenewLease(ctx, leaseID, azblob.ModifiedAccessConditions{})
	return err
}

func (s *azureBlob) AcquireLease(ctx context.Context, blobURL azblob.BlockBlobURL) (string, error) {
	const op errors.Op = "azureblob.AcquireLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	tctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()
	for {
		_, err := blobURL.Upload(tctx, bytes.NewReader([]byte{1}), azblob.BlobHTTPHeaders{}, nil, azblob.BlobAccessConditions{})
		if err != nil {
			if stgErr, ok := err.(azblob.StorageError); ok && stgErr.Response().StatusCode == http.StatusPreconditionFailed {
				select {
				case <-time.After(1 * time.Second):
					continue
				case <-tctx.Done():
					return "", tctx.Err()
				}
			}
			return "", errors.E(op, err)
		}

		leaseID, err := uuid.NewRandom()
		if err != nil {
			return "", errors.E(op, err)
		}
		res, err := blobURL.AcquireLease(tctx, leaseID.String(), 15, azblob.ModifiedAccessConditions{})
		if err != nil {
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
