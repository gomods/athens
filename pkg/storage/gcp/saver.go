package gcp

import (
	"context"
	"io"
	"log"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"

	moduploader "github.com/gomods/athens/pkg/storage/module"
)

// Save uploads the module's .mod, .zip and .info files for a given version
// It expects a context, which can be provided using context.Background
// from the standard library until context has been threaded down the stack.
// see issue: https://github.com/gomods/athens/issues/174
//
// Uploaded files are publicly accessible in the storage bucket as per
// an ACL rule.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "gcp.Save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	exists, err := s.Exists(ctx, module, version)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	if exists {
		return errors.E(op, "already exists", errors.M(module), errors.V(version), errors.KindAlreadyExists)
	}

	// err = moduploader.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.upload, s.timeout)
	err = moduploader.Upload(ctx, module, version, moduploader.NewStreamFromBytes(info), moduploader.NewStreamFromBytes(mod), moduploader.NewStreamFromReader(zip), s.upload, s.timeout)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded
	return nil
}

func (s *Storage) upload(ctx context.Context, path, contentType string, stream moduploader.Stream) error {
	const op errors.Op = "gcp.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	wc := s.bucket.Object(path).NewWriter(ctx)
	defer func(wc io.WriteCloser) {
		if err := wc.Close(); err != nil {
			log.Printf("WARNING: failed to close storage object writer: %s", err)
		}
	}(wc)
	// NOTE: content type is auto detected on GCP side and ACL defaults to public
	// Once we support private storage buckets this may need refactoring
	// unless there is a way to set the default perms in the project.
	if _, err := io.Copy(wc, stream.Stream); err != nil {
		return err
	}
	return nil
}
