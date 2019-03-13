package gcp

import (
	"bytes"
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	googleapi "google.golang.org/api/googleapi"
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
	gomodPath := config.PackageVersionedName(module, version, "mod")
	err := s.upload(ctx, gomodPath, bytes.NewReader(mod))
	if err != nil {
		return errors.E(op, err)
	}
	zipPath := config.PackageVersionedName(module, version, "zip")
	err = s.upload(ctx, zipPath, zip)
	if err != nil {
		return errors.E(op, err)
	}
	infoPath := config.PackageVersionedName(module, version, "info")
	err = s.upload(ctx, infoPath, bytes.NewReader(info))
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Storage) upload(ctx context.Context, path string, stream io.Reader) error {
	const op errors.Op = "gcp.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	wc := s.bucket.Object(path).If(storage.Conditions{
		DoesNotExist: true,
	}).NewWriter(ctx)

	// NOTE: content type is auto detected on GCP side and ACL defaults to public
	// Once we support private storage buckets this may need refactoring
	// unless there is a way to set the default perms in the project.
	if _, err := io.Copy(wc, stream); err != nil {
		wc.Close()
		return err
	}

	err := wc.Close()
	if err != nil {
		kind := errors.KindBadRequest
		apiErr, ok := err.(*googleapi.Error)
		if ok && apiErr.Code == 412 {
			kind = errors.KindAlreadyExists
		}
		return errors.E(op, err, kind)
	}
	return nil
}
