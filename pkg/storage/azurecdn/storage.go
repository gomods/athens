package azurecdn

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/gomods/athens/pkg/storage"
	"github.com/pkg/errors"
)

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module
type Storage struct {
	accountURL url.URL
	cred       blob.Credential
}

// New creates a new azure CDN saver
func New(accountName, accountKey string) (*Storage, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	if err != nil {
		return nil, err
	}
	cred := blob.NewSharedKeyCredential(accountName, accoutnKey)
	return &storage{accountURL: u, cred: cred}
}

// BaseURL returns the base URL that stores all modules. It can be used
// in the "meta" tag redirect response to vgo.
//
// For example:
//
//	<meta name="go-import" content="gomods.com/athens mod BaseURL()">
func (s *Storage) BaseURL() url.URL {
	return s.accountURL
}

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(module, version string, mod, zip, info []byte) error {
	ctx := context.Background()

	pipe := azblob.NewPipeline(s.cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(s.accountURL, pipe)
	// according to the first example in:
	// https://godoc.org/github.com/Azure/azure-storage-blob-go/2017-07-29/azblob
	// ... container names need to be lower case
	containerName := strings.ToLower(fmt.Sprintf("%s/@v", module))
	containerURL := serviceURL.NewContainerURL(containerName)

	// if the module already exists, the container will already exist.
	// this will be the case most of the time
	_, err := containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		// TODO: log that the container already exists
	}

	infoBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s.info"))
	modBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s.mod", version))
	zipBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s.zip", version))

	httpHeaders := func(contentType string) azblob.BlobHTTPHeaders {
		return azblob.BlobHTTPHeaders{
			ContentType: contentType
		}
	)
	// TODO: check errors
	if _, err := infoBlobURL.Upload(ctx, bytes.NewBuffer(info), httpHeaders("application/json")); err != nil {
		// TODO: log
		return err
	}
	if _, err := modBlobURL.Upload(ctx, bytes.NewBuffer(info), httpHeaders("text/plain")); err != nil {
		// TODO: log
		return err
	}
	if _, err := zipBlobURL.Upload(ctx, bytes.NewBuffer(zip), httpHeaders("application/octet-stream")); err != nil {
		// TODO: log
		return err
	}

	// TODO: take out lease on the /list file and add the version to it
}
