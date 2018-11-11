package azureblob

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	moduploader "github.com/gomods/athens/pkg/storage/module"
)

// Save implements the (github.com/gomods/athens/pkg/storage).Saver interface.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	fmt.Printf("Saving %v\n", module)
	const op errors.Op = "azureblob.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	err := moduploader.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.cl.UploadWithContext, s.conf.TimeoutDuration())
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	return nil
}
