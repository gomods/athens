package azureblob

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

type client interface {
	UploadWithContext(ctx context.Context, path, contentType string, content io.Reader) error
	BlobExists(ctx context.Context, path string) (bool, error)
	ReadBlob(ctx context.Context, path string) (io.ReadCloser, error)
}

type azureBlobStoreClient struct {
	containerURL *azblob.ContainerURL
}

func newBlobStoreClient(accountURL *url.URL, accountName, accountKey, containerName string) *azureBlobStoreClient {
	cred := azblob.NewSharedKeyCredential(accountName, accountKey)
	pipe := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*accountURL, pipe)
	// rules on container names:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#container-names
	//
	// This container must exist
	containerURL := serviceURL.NewContainerURL(containerName)
	cl := &azureBlobStoreClient{containerURL: &containerURL}
	return cl
}

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module
type Storage struct {
	cl   client
	conf *config.AzureBlobConfig
}

// New creates a new azure blobs storage saver
func New(conf *config.AzureBlobConfig) (*Storage, error) {
	const op errors.Op = "azure.New"
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	cl := newBlobStoreClient(u, conf.AccountName, conf.AccountKey, conf.ContainerName)
	return &Storage{cl: cl, conf: conf}, nil
}

// BlobExists checks if a particular blob exists in the container
func (c *azureBlobStoreClient) BlobExists(ctx context.Context, path string) (bool, error) {
	// TODO: Any better way of doing this ?
	blobURL := c.containerURL.NewBlockBlobURL(path)
	_, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{})
	if err != nil {
		serr := err.(azblob.StorageError)
		if serr.Response().StatusCode == 404 {
			return false, nil
		}

		return false, err
	}
	return true, nil

}

// ReadBlob returns an io.ReadCloser for the contents of a blob
func (c *azureBlobStoreClient) ReadBlob(ctx context.Context, path string) (io.ReadCloser, error) {
	blobURL := c.containerURL.NewBlockBlobURL(path)
	downloadResponse, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return nil, err
	}
	return downloadResponse.Body(azblob.RetryReaderOptions{}), nil
}

// UploadWithContext uploads a blob to the container
func (c *azureBlobStoreClient) UploadWithContext(ctx context.Context, path, contentType string, content io.Reader) error {
	const op errors.Op = "azure.UploadWithContext"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	blobURL := c.containerURL.NewBlockBlobURL(path)
	emptyMeta := map[string]string{}
	emptyBlobAccessCond := azblob.BlobAccessConditions{}
	httpHeaders := func(contentType string) azblob.BlobHTTPHeaders {
		return azblob.BlobHTTPHeaders{
			ContentType: contentType,
		}
	}
	bufferSize := 1 * 1024 * 1024 // Size of the rotating buffers that are used when uploading
	maxBuffers := 3               // Number of rotating buffers that are used when uploading

	uploadStreamOpts := azblob.UploadStreamToBlockBlobOptions{
		BufferSize:       bufferSize,
		MaxBuffers:       maxBuffers,
		BlobHTTPHeaders:  httpHeaders(contentType),
		Metadata:         emptyMeta,
		AccessConditions: emptyBlobAccessCond,
	}
	_, err := azblob.UploadStreamToBlockBlob(ctx, content, blobURL, uploadStreamOpts)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}
