package azureblob

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

type azureBlobStoreClient struct {
	containerURL *azblob.ContainerURL
}

const (
	// TokenRefreshTolerance defines the duration before the token's actual expiration time
	// during which the token should be refreshed. This helps ensure that the token is
	// refreshed in a timely manner, avoiding potential issues with token expiration.
	TokenRefreshTolerance = 5 * time.Minute
)

func newBlobStoreClient(accountURL *url.URL, accountName, accountKey, credScope, managedIdentityResourceID, containerName string) (*azureBlobStoreClient, error) {
	const op errors.Op = "azureblob.newBlobStoreClient"
	var pipe pipeline.Pipeline
	if managedIdentityResourceID != "" {
		msiCred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ResourceID(managedIdentityResourceID),
		})
		if err != nil {
			return nil, errors.E(op, err)
		}
		token, err := msiCred.GetToken(context.Background(), policy.TokenRequestOptions{
			Scopes: []string{credScope},
		})
		if err != nil {
			return nil, errors.E(op, err)
		}
		tokenCred := azblob.NewTokenCredential(token.Token, func(tc azblob.TokenCredential) time.Duration {
			fmt.Printf("refreshing token started at: %s", time.Now())
			refreshedToken, err := msiCred.GetToken(context.Background(), policy.TokenRequestOptions{
				Scopes: []string{credScope},
			})
			if err != nil {
				fmt.Printf("error getting token: %s during token refresh process", err)
				// token refresh may fail due to transient errors, so we return a non-zero duration
				// to retry the token refresh after a short delay
				return time.Minute
			}
			tc.SetToken(refreshedToken.Token)

			refreshDuration := time.Until(refreshedToken.ExpiresOn.Add(-TokenRefreshTolerance))
			fmt.Printf("refresh duration: %s", refreshDuration)
			return refreshDuration
		})
		pipe = azblob.NewPipeline(tokenCred, azblob.PipelineOptions{})
	}
	if pipe == nil && accountKey != "" {
		cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return nil, errors.E(op, err)
		}
		pipe = azblob.NewPipeline(cred, azblob.PipelineOptions{})
	}
	serviceURL := azblob.NewServiceURL(*accountURL, pipe)
	// rules on container names:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#container-names
	//
	// This container must exist
	containerURL := serviceURL.NewContainerURL(containerName)
	cl := &azureBlobStoreClient{containerURL: &containerURL}
	return cl, nil
}

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module.
type Storage struct {
	client  *azureBlobStoreClient
	timeout time.Duration
}

// New creates a new azure blobs storage.
func New(conf *config.AzureBlobConfig, timeout time.Duration) (*Storage, error) {
	const op errors.Op = "azureblob.New"
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	if conf.AccountKey == "" && (conf.ManagedIdentityResourceID == "" || conf.CredentialScope == "") {
		return nil, errors.E(op, "either account key or managed identity resource id and storage resource must be set")
	}
	cl, err := newBlobStoreClient(u, conf.AccountName, conf.AccountKey, conf.CredentialScope, conf.ManagedIdentityResourceID, conf.ContainerName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return &Storage{client: cl, timeout: timeout}, nil
}

// BlobExists checks if a particular blob exists in the container.
func (c *azureBlobStoreClient) BlobExists(ctx context.Context, path string) (bool, error) {
	const op errors.Op = "azureblob.BlobExists"
	// TODO: Any better way of doing this ?
	blobURL := c.containerURL.NewBlockBlobURL(path)
	_, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{})
	if err != nil {
		var serr azblob.StorageError
		if !errors.AsErr(err, &serr) {
			return false, errors.E(op, fmt.Errorf("error in casting to azure error type %w", err))
		}
		if serr.Response().StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, errors.E(op, err)
	}
	return true, nil
}

// ReadBlob returns a storage.SizeReadCloser for the contents of a blob.
func (c *azureBlobStoreClient) ReadBlob(ctx context.Context, path string) (storage.SizeReadCloser, error) {
	const op errors.Op = "azureblob.ReadBlob"
	blobURL := c.containerURL.NewBlockBlobURL(path)
	downloadResponse, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, errors.E(op, err)
	}
	rc := downloadResponse.Body(azblob.RetryReaderOptions{})
	size := downloadResponse.ContentLength()
	return storage.NewSizer(rc, size), nil
}

// ListBlobs will list all blobs which has the given prefix.
func (c *azureBlobStoreClient) ListBlobs(ctx context.Context, prefix string) ([]string, error) {
	const op errors.Op = "azureblob.ListBlobs"
	var blobs []string
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := c.containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{
			Prefix: prefix,
		})
		if err != nil {
			return nil, errors.E(op, err)
		}
		marker = listBlob.NextMarker

		for _, blob := range listBlob.Segment.BlobItems {
			blobs = append(blobs, blob.Name)
		}
	}

	return blobs, nil
}

// DeleteBlob deletes the blob with the given path.
func (c *azureBlobStoreClient) DeleteBlob(ctx context.Context, path string) error {
	const op errors.Op = "azureblob.DeleteBlob"
	blobURL := c.containerURL.NewBlockBlobURL(path)
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// UploadWithContext uploads a blob to the container.
func (c *azureBlobStoreClient) UploadWithContext(ctx context.Context, path, contentType string, content io.Reader) error {
	const op errors.Op = "azureblob.UploadWithContext"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	blobURL := c.containerURL.NewBlockBlobURL(path)
	bufferSize := 1 * 1024 * 1024 // Size of the rotating buffers that are used when uploading
	maxBuffers := 3               // Number of rotating buffers that are used when uploading

	uploadStreamOpts := azblob.UploadStreamToBlockBlobOptions{
		BufferSize: bufferSize,
		MaxBuffers: maxBuffers,
		BlobHTTPHeaders: azblob.BlobHTTPHeaders{
			ContentType: contentType,
		},
	}
	_, err := azblob.UploadStreamToBlockBlob(ctx, content, blobURL, uploadStreamOpts)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}
