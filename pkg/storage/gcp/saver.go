package gcp

import (
	"bytes"
	"context"
	"io"
	"time"

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
	const op errors.Op = "gcp.save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	err := s.save(ctx, module, version, mod, zip, info)
	if err != nil {
		return errors.E(op, err)
	}
	return err
}

// SetStaleThreshold sets the threshold of how long we consider
// a lock metadata stale after.
func (s *Storage) SetStaleThreshold(threshold time.Duration) {
	s.staleThreshold = threshold
}

func (s *Storage) save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "gcp.save"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	gomodPath := config.PackageVersionedName(module, version, "mod")
	err := s.upload(ctx, gomodPath, bytes.NewReader(mod), false)
	// KindAlreadyExists means the file is uploaded (somewhere else) successfully.
	if err != nil && !errors.Is(err, errors.KindAlreadyExists) {
		return errors.E(op, err)
	}

	zipPath := config.PackageVersionedName(module, version, "zip")
	err = s.upload(ctx, zipPath, zip, true)
	if err != nil && !errors.Is(err, errors.KindAlreadyExists) {
		return errors.E(op, err)
	}

	infoPath := config.PackageVersionedName(module, version, "info")
	err = s.upload(ctx, infoPath, bytes.NewReader(info), false)
	if err != nil && !errors.Is(err, errors.KindAlreadyExists) {
		return errors.E(op, err)
	}

	return nil
}

func (s *Storage) upload(ctx context.Context, path string, stream io.Reader, checkBefore bool) error {
	const op errors.Op = "gcp.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if checkBefore {
		// Check whether the file already exists before uploading.
		// Note that this is not for preventing the same file from being uploaded multiple times,
		// but only a small optimization to avoid unnecessary uploads for large files (in particular .zip file).
		_, err := s.bucket.Object(path).Attrs(cancelCtx)
		if err == nil {
			// The file already exists, no need to upload it again.
			return nil
		} else if !errors.IsErr(err, storage.ErrObjectNotExist) {
			// Not expected error, return it.
			return errors.E(op, err)
		}
		// Otherwise, the error is ErrObjectNotExist, so we should upload the file.
	}

	wc := s.bucket.Object(path).If(storage.Conditions{
		DoesNotExist: true,
	}).NewWriter(cancelCtx)

	// NOTE: content type is auto detected on GCP side and ACL defaults to public
	// Once we support private storage buckets this may need refactoring
	// unless there is a way to set the default perms in the project.
	if _, err := io.Copy(wc, stream); err != nil {
		// Purposely do not close it to avoid creating a partial file.
		return err
	}

	err := wc.Close()
	if err != nil {
		kind := errors.KindBadRequest
		apiErr := &googleapi.Error{}
		if errors.AsErr(err, &apiErr) && apiErr.Code == 412 {
			kind = errors.KindAlreadyExists
		}
		return errors.E(op, err, kind)
	}
	return nil
}
