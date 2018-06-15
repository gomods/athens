package azurecdn

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/bketelsen/buffet"
	"github.com/gobuffalo/buffalo"
)

// Storage implements (github.com/gomods/athens/pkg/storage).Saver and
// also provides a function to fetch the location of a module
type Storage struct {
	accountURL *url.URL
	cred       azblob.Credential
}

// New creates a new azure CDN saver
func New(accountName, accountKey string) (*Storage, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	if err != nil {
		return nil, err
	}
	cred := azblob.NewSharedKeyCredential(accountName, accountKey)
	return &Storage{accountURL: u, cred: cred}, nil
}

// BaseURL returns the base URL that stores all modules. It can be used
// in the "meta" tag redirect response to vgo.
//
// For example:
//
//	<meta name="go-import" content="gomods.com/athens mod BaseURL()">
func (s Storage) BaseURL() *url.URL {
	return s.accountURL
}

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(c buffalo.Context, module, version string, mod, zip, info []byte) error {
	sp := buffet.ChildSpan("storage.save", c)
	defer sp.Finish()

	pipe := azblob.NewPipeline(s.cred, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*s.accountURL, pipe)
	// rules on container names:
	// https://docs.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#container-names
	//
	// This container must exist
	containerURL := serviceURL.NewContainerURL("gomodules")

	infoBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s/@v/%s.info", module, version))
	modBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s/@v/%s.mod", module, version))
	zipBlobURL := containerURL.NewBlockBlobURL(fmt.Sprintf("%s/@v/%s.zip", module, version))

	httpHeaders := func(contentType string) azblob.BlobHTTPHeaders {
		return azblob.BlobHTTPHeaders{
			ContentType: contentType,
		}
	}
	emptyMeta := map[string]string{}
	emptyBlobAccessCond := azblob.BlobAccessConditions{}

	infoErr := make(chan error)
	go func(errOut chan<- error) {
		defer close(errOut)
		sp := buffet.ChildSpan("storage.save.info", c)
		defer sp.Finish()

		_, err := infoBlobURL.Upload(c, bytes.NewReader(info), httpHeaders("application/json"), emptyMeta, emptyBlobAccessCond)
		errOut <- err
	}(infoErr)

	modErr := make(chan error)
	go func(errOut chan<- error) {
		defer close(errOut)
		sp := buffet.ChildSpan("storage.save.module", c)
		defer sp.Finish()

		_, err := modBlobURL.Upload(c, bytes.NewReader(info), httpHeaders("text/plain"), emptyMeta, emptyBlobAccessCond)
		errOut <- err
	}(modErr)

	zipErr := make(chan error)
	go func(errOut chan<- error) {
		defer close(errOut)
		sp := buffet.ChildSpan("storage.save.zip", c)
		defer sp.Finish()

		_, err := zipBlobURL.Upload(c, bytes.NewReader(zip), httpHeaders("application/octet-stream"), emptyMeta, emptyBlobAccessCond)
		errOut <- err
	}(zipErr)

	select {
	case err := <-infoErr:
		if err != nil {
			return err
		}
		// TODO: log
	case <-c.Done():
		return c.Err()
	}

	select {
	case err := <-modErr:
		if err != nil {
			return err
		}
		// TODO: log
	case <-c.Done():
		return c.Err()
	}

	select {
	case err := <-zipErr:
		if err != nil {
			return err
		}
		// TODO: log
	case <-c.Done():
		return c.Err()
	}

	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded

	return nil
}
