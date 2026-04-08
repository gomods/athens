package azureblob

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

type azureBlobStoreClient struct {
	client        *azblob.Client
	containerName string
}

func newBlobStoreClient(serviceURL, accountName, accountKey, managedIdentityResourceID, containerName string) (*azureBlobStoreClient, error) {
	const op errors.Op = "azureblob.newBlobStoreClient"

	var client *azblob.Client

	if managedIdentityResourceID != "" {
		msiCred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ResourceID(managedIdentityResourceID),
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

	if client == nil && accountKey != "" {
		cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return nil, errors.E(op, err)
		}

		c, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, errors.E(op, err)
		}

		client = c
	}

	cl := &azureBlobStoreClient{client: client, containerName: containerName}

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

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName)

	if conf.AccountKey == "" && (conf.ManagedIdentityResourceID == "" || conf.CredentialScope == "") {
		return nil, errors.E(op, "either account key or managed identity resource id and storage resource must be set")
	}

	cl, err := newBlobStoreClient(serviceURL, conf.AccountName, conf.AccountKey, conf.ManagedIdentityResourceID, conf.ContainerName)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &Storage{client: cl, timeout: timeout}, nil
}

// BlobExists checks if a particular blob exists in the container.
func (c *azureBlobStoreClient) BlobExists(ctx context.Context, path string) (bool, error) {
	const op errors.Op = "azureblob.BlobExists"

	blobClient := c.client.ServiceClient().NewContainerClient(c.containerName).NewBlockBlobClient(path)

	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if !errors.AsErr(err, &respErr) {
			return false, errors.E(op, fmt.Errorf("error in casting to azure error type %w", err))
		}

		if respErr.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, errors.E(op, err)
	}

	return true, nil
}

// ReadBlob returns a storage.SizeReadCloser for the contents of a blob.
func (c *azureBlobStoreClient) ReadBlob(ctx context.Context, path string) (storage.SizeReadCloser, error) {
	const op errors.Op = "azureblob.ReadBlob"

	resp, err := c.client.DownloadStream(ctx, c.containerName, path, nil)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var size int64
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}

	return storage.NewSizer(resp.Body, size), nil
}

// ListBlobs will list all blobs which has the given prefix.
func (c *azureBlobStoreClient) ListBlobs(ctx context.Context, prefix string) ([]string, error) {
	const op errors.Op = "azureblob.ListBlobs"

	var blobs []string

	pager := c.client.NewListBlobsFlatPager(c.containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.E(op, err)
		}

		for _, item := range resp.Segment.BlobItems {
			blobs = append(blobs, *item.Name)
		}
	}

	return blobs, nil
}

// DeleteBlob deletes the blob with the given path.
func (c *azureBlobStoreClient) DeleteBlob(ctx context.Context, path string) error {
	const op errors.Op = "azureblob.DeleteBlob"

	_, err := c.client.DeleteBlob(ctx, c.containerName, path, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return errors.E(op, err, errors.KindNotFound)
		}

		return errors.E(op, err)
	}

	return nil
}

// UploadWithContext uploads a blob to the container.
func (c *azureBlobStoreClient) UploadWithContext(ctx context.Context, path, contentType string, content io.Reader) error {
	const op errors.Op = "azureblob.UploadWithContext"

	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	bufferSize := 1 * 1024 * 1024 // Size of the rotating buffers that are used when uploading
	maxBuffers := 3               // Number of rotating buffers that are used when uploading

	_, err := c.client.UploadStream(ctx, c.containerName, path, content, &azblob.UploadStreamOptions{
		BlockSize:   int64(bufferSize),
		Concurrency: maxBuffers,
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
