package stash

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/azureblob"
	"github.com/google/uuid"
)

// WithAzureBlobLock returns a distributed singleflight
// using a Azure Blob Storage backend. See the config.toml documentation for details.
func WithAzureBlobLock(conf *config.AzureBlobConfig, timeout time.Duration, checker storage.Checker) (Wrapper, error) {
	const op errors.Op = "stash.WithAzureBlobLock"

	if conf.AccountKey == "" && (conf.ManagedIdentityResourceID == "" || conf.CredentialScope == "") {
		return nil, errors.E(op, "either account key or managed identity resource id and storage resource must be set")
	}
	accountURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	var cred azblob.Credential
	if conf.AccountKey != "" {
		cred, err = azblob.NewSharedKeyCredential(conf.AccountName, conf.AccountKey)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}
	if conf.ManagedIdentityResourceID != "" {
		msiCred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ResourceID(conf.ManagedIdentityResourceID),
		})
		if err != nil {
			return nil, errors.E(op, err)
		}
		token, err := msiCred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{conf.CredentialScope}})
		if err != nil {
			return nil, errors.E(op, err)
		}
		cred = azblob.NewTokenCredential(token.Token, func(tc azblob.TokenCredential) time.Duration {
			fmt.Printf("refreshing token started at: %s", time.Now())
			refreshedToken, err := msiCred.GetToken(context.Background(), policy.TokenRequestOptions{
				Scopes: []string{conf.CredentialScope},
			})
			if err != nil {
				fmt.Printf("error getting token: %s during token refresh process", err)
				// token refresh may fail due to transient errors, so we return a non-zero duration
				// to retry the token refresh after a short delay
				return time.Minute
			}
			tc.SetToken(refreshedToken.Token)

			refreshDuration := time.Until(refreshedToken.ExpiresOn.Add(-azureblob.TokenRefreshTolerance))
			fmt.Printf("refresh duration: %s", refreshDuration)
			return refreshDuration
		})
	}
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*accountURL, pipe)

	containerURL := serviceURL.NewContainerURL(conf.ContainerName)

	return func(s Stasher) Stasher {
		return &azblobLock{containerURL, s, checker}
	}, nil
}

type azblobLock struct {
	containerURL azblob.ContainerURL
	stasher      Stasher
	checker      storage.Checker
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
	leaseBlobURL := s.containerURL.NewBlockBlobURL(leaseBlobName)

	leaseID, err := s.acquireLease(ctx, leaseBlobURL)
	if err != nil {
		return ver, errors.E(op, err)
	}
	defer func() {
		const op errors.Op = "azblobLock.Unlock"
		relErr := s.releaseLease(ctx, leaseBlobURL, leaseID)
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
			err := s.renewLease(ctx, leaseBlobURL, leaseID)
			if err != nil {
				return ver, errors.E(op, err)
			}
		case <-ctx.Done():
			return ver, errors.E(op, ctx.Err())
		}
	}
}

func (s *azblobLock) releaseLease(ctx context.Context, blobURL azblob.BlockBlobURL, leaseID string) error {
	const op errors.Op = "azblobLock.releaseLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := blobURL.ReleaseLease(ctx, leaseID, azblob.ModifiedAccessConditions{})
	return err
}

func (s *azblobLock) renewLease(ctx context.Context, blobURL azblob.BlockBlobURL, leaseID string) error {
	const op errors.Op = "azblobLock.renewLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := blobURL.RenewLease(ctx, leaseID, azblob.ModifiedAccessConditions{})
	return err
}

func (s *azblobLock) acquireLease(ctx context.Context, blobURL azblob.BlockBlobURL) (string, error) {
	const op errors.Op = "azblobLock.acquireLease"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	tctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// first we need to create a blob which can be then leased
	_, err := blobURL.Upload(tctx, bytes.NewReader([]byte{1}), azblob.BlobHTTPHeaders{}, nil, azblob.BlobAccessConditions{})
	if err != nil {
		// if the blob is already leased we will get http.StatusPreconditionFailed while writing to that blob
		var stgErr azblob.StorageError
		if !errors.AsErr(err, &stgErr) || stgErr.Response().StatusCode != http.StatusPreconditionFailed {
			return "", errors.E(op, err)
		}
	}

	leaseID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.E(op, err)
	}
	for {
		// acquire lease for 15 sec (it's the min value)
		res, err := blobURL.AcquireLease(tctx, leaseID.String(), 15, azblob.ModifiedAccessConditions{})
		if err != nil {
			// if the blob is already leased we will get http.StatusConflict - wait and try again
			var stgErr azblob.StorageError
			if ok := errors.AsErr(err, &stgErr); ok && stgErr.Response().StatusCode == http.StatusConflict {
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
