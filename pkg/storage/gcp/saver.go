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

// Fallback for how long we consider an "in_progress" metadata key stale,
// due to failure to remove it.
const fallbackInProgressStaleThreshold = 2 * time.Minute

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
	gomodPath := config.PackageVersionedName(module, version, "mod")
	innerErr := s.save(ctx, module, version, mod, zip, info)
	if errors.Is(innerErr, errors.KindAlreadyExists) {
		// Cache hit.
		return errors.E(op, innerErr)
	}
	// No cache hit. Remove the metadata lock if it is there.
	inProgress, outerErr := s.checkUploadInProgress(ctx, gomodPath)
	if outerErr != nil {
		return errors.E(op, outerErr)
	}
	if inProgress {
		outerErr = s.removeInProgressMetadata(ctx, gomodPath)
		if outerErr != nil {
			return errors.E(op, outerErr)
		}
	}
	return innerErr
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
	seenAlreadyExists := 0
	err := s.upload(ctx, gomodPath, bytes.NewReader(mod), true)
	// If it already exists, check the object metadata to see if the
	// other two are still uploading in progress somewhere else. If they
	// are, return a cache hit. If not, continue on to the other two,
	// and only return a cache hit if all three exist.
	if errors.Is(err, errors.KindAlreadyExists) {
		inProgress, progressErr := s.checkUploadInProgress(ctx, gomodPath)
		if progressErr != nil {
			return errors.E(op, progressErr)
		}
		if inProgress {
			// err is known to be errors.KindAlreadyExists at this point, so
			// this is a cache hit return.
			return errors.E(op, err)
		}
		seenAlreadyExists++
	} else if err != nil {
		// Other errors
		return errors.E(op, err)
	}
	zipPath := config.PackageVersionedName(module, version, "zip")
	err = s.upload(ctx, zipPath, zip, false)
	if errors.Is(err, errors.KindAlreadyExists) {
		seenAlreadyExists++
	} else if err != nil {
		return errors.E(op, err)
	}
	infoPath := config.PackageVersionedName(module, version, "info")
	err = s.upload(ctx, infoPath, bytes.NewReader(info), false)
	// Have all three returned errors.KindAlreadyExists?
	if errors.Is(err, errors.KindAlreadyExists) {
		if seenAlreadyExists == 2 {
			return errors.E(op, err)
		}
	} else if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Storage) removeInProgressMetadata(ctx context.Context, gomodPath string) error {
	const op errors.Op = "gcp.removeInProgressMetadata"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	_, err := s.bucket.Object(gomodPath).Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{},
	})
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Storage) checkUploadInProgress(ctx context.Context, gomodPath string) (bool, error) {
	const op errors.Op = "gcp.checkUploadInProgress"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	attrs, err := s.bucket.Object(gomodPath).Attrs(ctx)
	if err != nil {
		return false, errors.E(op, err)
	}
	// If we have a config-set lock threshold, i.e. we are using the GCP
	// slightflight backend, use it. Otherwise, use the fallback, which
	// is arguably irrelevant when not using GCP for singleflighting.
	threshold := fallbackInProgressStaleThreshold
	if s.staleThreshold > 0 {
		threshold = s.staleThreshold
	}
	if attrs.Metadata != nil {
		_, ok := attrs.Metadata["in_progress"]
		if ok {
			// In case the final call to remove the metadata fails for some reason,
			// we have a threshold after which we consider this to be stale.
			if time.Since(attrs.Created) > threshold {
				return false, nil
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *Storage) upload(ctx context.Context, path string, stream io.Reader, first bool) error {
	const op errors.Op = "gcp.upload"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wc := s.bucket.Object(path).If(storage.Conditions{
		DoesNotExist: true,
	}).NewWriter(cancelCtx)

	// We set this metadata only for the first of the three files uploaded,
	// for use as a singleflight lock.
	if first {
		wc.ObjectAttrs.Metadata = make(map[string]string)
		wc.ObjectAttrs.Metadata["in_progress"] = "true"
	}

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
