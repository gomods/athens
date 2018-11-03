package azureblob

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
)

type client interface {
	UploadWithContext(ctx context.Context, path, contentType string, content io.Reader) error
}

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module
type Storage struct {
	cl      client
	baseURI *url.URL
	cdnConf *config.CDNConfig
}

// New creates a new azureblob saver
func New(azblobConf config.AzureBlobConf, cdnConf *config.CDNConfig) (*Storage, error) {
	const op errors.Op = "azureblob.New"
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", azblobConf.AccountName))
	if err != nil {
		return nil, errors.E(op, err)
	}
	cl := newBlobStoreClient(u, azblobConf.AccountName, azblobConf.AccountKey, azblobConf.ContainerName)
	return &Storage{cl: cl, baseURI: u, cdnConf: cdnConf}, nil
}

// BaseURL returns the base URL that stores all modules. It can be used
// in the "meta" tag redirect response to vgo.
//
// For example:
//
//	<meta name="go-import" content="gomods.com/athens mod BaseURL()">
func (s Storage) BaseURL() *url.URL {
	return s.cdnConf.CDNEndpointWithDefault(s.baseURI)
}
