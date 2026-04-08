package stash

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/lease"
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

	if conf.AccountKey == "" && (conf.ManagedIdentityResourceID == "" || conf.CredentialScope == "") {
		return nil, errors.E(op, "either account key or managed identity resource id and storage resource must be set")
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName)

	var client *azblob.Client

	if conf.ManagedIdentityResourceID != "" {
		msiCred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ResourceID(conf.ManagedIdentityResourceID),
		})
		if err != nil {
			return nil, errors.E(op, err)
		}

		c, err := azblob.NewClient(serviceURL, msiCred, nil)
		if err != nil {
			return nil, errors.E(op, err)
		}

		client = c
	}

	if client == nil && conf.AccountKey != "" {
		cred, err := azblob.NewSharedKeyCredential(conf.AccountName, conf.AccountKey)
		if err != nil {
			return nil, errors.E(op, err)
		}

		c, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, errors.E(op, err)
		}

		client = c
	}

	containerName := conf.ContainerName

	return func(s Stasher) Stasher {
		return &azblobLock{client: client, containerName: containerName, stasher: s, checker: checker}
	}, nil
}

type azblobLock struct {
	client        *azblob.Client
	containerName string
	stasher       Stasher
	checker       storage.Checker
}

type stashRes struct {
	v   string
	err error
}

func (s *azblobLock) Stash(ctx context.Context, mod, ver string) (newVer string, err error) {
	const op errors.Op = "azblobLock.Stash"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	leaseBlobName := "lease/" + config.FmtModVer(mod, ver)

	leaseID, err := s.acquireLease(ctx, leaseBlobName)
	if err != nil {
		return ver, errors.E(op, err)
	}

	defer func() {
		const op errors.Op = "azblobLock.Unlock"

		relErr := s.releaseLease(ctx, leaseBlobName, leaseID)
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

	sChan := make(chan stashRes)

	go func() {
		v, err := s.stasher.Stash(ctx, mod, ver)
		sChan <- stashRes{v, err}
	}()

	for {
		select {
		case sr := <-sChan:
			if sr.err != nil {
				err = errors.E(op, sr.err)

				return ver, err
			}

			newVer = sr.v

			return newVer, nil
		case <-time.After(10 * time.Second):
			err := s.renewLease(ctx, leaseBlobName, leaseID)
			if err != nil {
				return ver, errors.E(op, err)
			}
		case <-ctx.Done():
			return ver, errors.E(op, ctx.Err())
		}
	}
}

func (s *azblobLock) blobLeaseClient(blobName, leaseID string) (*lease.BlobClient, error) {
	containerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	blobClient := containerClient.NewBlobClient(blobName)

	return lease.NewBlobClient(blobClient, &lease.BlobClientOptions{
		LeaseID: &leaseID,
	})
}

func (s *azblobLock) releaseLease(ctx context.Context, blobName, leaseID string) error {
	const op errors.Op = "azblobLock.releaseLease"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	leaseClient, err := s.blobLeaseClient(blobName, leaseID)
	if err != nil {
		return err
	}

	_, err = leaseClient.ReleaseLease(ctx, nil)

	return err
}

func (s *azblobLock) renewLease(ctx context.Context, blobName, leaseID string) error {
	const op errors.Op = "azblobLock.renewLease"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	leaseClient, err := s.blobLeaseClient(blobName, leaseID)
	if err != nil {
		return err
	}

	_, err = leaseClient.RenewLease(ctx, nil)

	return err
}

func (s *azblobLock) acquireLease(ctx context.Context, blobName string) (string, error) {
	const op errors.Op = "azblobLock.acquireLease"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	tctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// first we need to create a blob which can be then leased
	containerClient := s.client.ServiceClient().NewContainerClient(s.containerName)
	bbClient := containerClient.NewBlockBlobClient(blobName)

	_, err := bbClient.Upload(tctx, streaming.NopCloser(bytes.NewReader([]byte{1})), nil)
	if err != nil {
		// if the blob is already leased we will get http.StatusPreconditionFailed while writing to that blob
		var respErr *azcore.ResponseError
		if !errors.AsErr(err, &respErr) || respErr.StatusCode != http.StatusPreconditionFailed {
			return "", errors.E(op, err)
		}
	}

	leaseID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.E(op, err)
	}

	leaseIDStr := leaseID.String()

	for {
		leaseClient, err := s.blobLeaseClient(blobName, leaseIDStr)
		if err != nil {
			return "", errors.E(op, err)
		}

		// acquire lease for 15 sec (it's the min value)
		duration := int32(15)

		_, err = leaseClient.AcquireLease(tctx, duration, nil)
		if err != nil {
			// if the blob is already leased we will get http.StatusConflict - wait and try again
			var respErr *azcore.ResponseError
			if ok := errors.AsErr(err, &respErr); ok && respErr.StatusCode == http.StatusConflict {
				select {
				case <-time.After(1 * time.Second):
					continue
				case <-tctx.Done():
					return "", tctx.Err()
				}
			}

			return "", errors.E(op, err)
		}

		return leaseIDStr, nil
	}
}
