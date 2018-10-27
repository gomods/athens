package multi

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/gomods/athens/pkg/storage"
)

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "mutli.Delete"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	content, err := ioutil.ReadAll(zip)
	if err != nil {
		return err
	}

	var results = make(chan error, len(s.storages))
	c, cancel := context.WithCancel(ctx)

	for _, store := range s.storages {
		go func(sb storage.Backend) {
			rdr := bytes.NewReader(content)
			select {
			case results <- sb.Save(c, module, version, mod, rdr, info):
			case <-c.Done():
			}
		}(store)
	}

	var errs []error
	for i := 0; i < len(s.storages); i++ {
		r := <-results
		if r != nil {
			errs = append(errs, r)
		}
	}

	cancel()
	close(results)

	return s.composeError(module, version, op, errs...)
}
