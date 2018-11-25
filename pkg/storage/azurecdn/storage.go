package azure

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	moduploader "github.com/gomods/athens/pkg/storage/module"
)

type client interface {
	UploadWithContext(ctx observ.ProxyContext, path, contentType string, content io.Reader) error
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
	conf *config.AzureConfig
}

// New creates a new azure blobs storage saver
func New(conf *config.AzureConfig) (*Storage, error) {
	const op errors.Op = "azure.New"
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	cl := newBlobStoreClient(u, conf.AccountName, conf.AccountKey, conf.ContainerName)
	return &Storage{cl: cl, conf: conf}, nil
}

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx observ.ProxyContext, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "azure.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	err := moduploader.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.cl.UploadWithContext, s.conf.TimeoutDuration())
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}

func (c *azureBlobStoreClient) UploadWithContext(ctx observ.ProxyContext, path, contentType string, content io.Reader) error {
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
