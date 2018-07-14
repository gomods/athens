package azurecdn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/gomods/athens/pkg/config/env"
	m "github.com/gomods/athens/pkg/storage/module"
)

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module
type Storage struct {
	accountURL    *url.URL
	cred          azblob.Credential
	containerName string
}

// New creates a new azure CDN saver
func New(accountName, accountKey, containerName string) (*Storage, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	if err != nil {
		return nil, err
	}
	cred := azblob.NewSharedKeyCredential(accountName, accountKey)
	return &Storage{accountURL: u, cred: cred, containerName: containerName}, nil
}

// BaseURL returns the base URL that stores all modules. It can be used
// in the "meta" tag redirect response to vgo.
//
// For example:
//
//	<meta name="go-import" content="gomods.com/athens mod BaseURL()">
func (s Storage) BaseURL() *url.URL {
	return env.CDNEndpointWithDefault(s.accountURL)
}

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	pipe := azblob.NewPipeline(s.cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*s.accountURL, pipe)
	// rules on container names:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#container-names
	//
	// This container must exist
	containerURL := serviceURL.NewContainerURL(s.containerName)
	err := m.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, getUploader(containerURL))
	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded
	return err
}

func getUploader(containerURL azblob.ContainerURL) m.Uploader {
	emptyMeta := map[string]string{}
	emptyBlobAccessCond := azblob.BlobAccessConditions{}
	httpHeaders := func(contentType string) azblob.BlobHTTPHeaders {
		return azblob.BlobHTTPHeaders{
			ContentType: contentType,
		}
	}

	return func(ctx context.Context, path, contentType string, stream io.Reader) error {
		// TODO: find out which values make sense here
		bufferSize := 1 * 1024 * 1024 // Size of the rotating buffers that are used when uploading
		maxBuffers := 3               // Number of rotating buffers that are used when uploading

		uploadStreamOpts := azblob.UploadStreamToBlockBlobOptions{
			BufferSize:       bufferSize,
			MaxBuffers:       maxBuffers,
			BlobHTTPHeaders:  httpHeaders(contentType),
			Metadata:         emptyMeta,
			AccessConditions: emptyBlobAccessCond,
		}

		url := containerURL.NewBlockBlobURL(path)
		_, err := azblob.UploadStreamToBlockBlob(ctx, stream, url, uploadStreamOpts)
		return err
	}
}
