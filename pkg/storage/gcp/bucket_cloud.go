package gcp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gomods/athens/pkg/config"
	s "github.com/gomods/athens/pkg/storage"
	multierror "github.com/hashicorp/go-multierror"
	"google.golang.org/api/iterator"
)

type gcpBucket struct {
	*storage.BucketHandle
}

func (b *gcpBucket) DeleteModule(ctx context.Context, module, version string) error {
	if !b.ObjectExists(ctx, module, version) {
		return s.ErrVersionNotFound{Module: module, Version: version}
	}
	var errs error
	if err := b.Object(config.PackageVersionedName(module, version, "mod")).Delete(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := b.Object(config.PackageVersionedName(module, version, "info")).Delete(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := b.Object(config.PackageVersionedName(module, version, "zip")).Delete(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	return errs
}

func (b *gcpBucket) GetReader(ctx context.Context, filename string) (io.ReadCloser, error) {
	return b.Object(filename).NewReader(ctx)
}

func (b *gcpBucket) GetWriter(ctx context.Context, filename string) io.WriteCloser {
	return b.Object(filename).NewWriter(ctx)
}

func (b *gcpBucket) ListVersions(ctx context.Context, module string) ([]string, error) {
	it := b.Objects(ctx, &storage.Query{Prefix: module})

	versions := make([]string, 0, 10)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not iterate over query: %s", err)
		}

		// kinda hacky looking at this time
		if strings.HasSuffix(attrs.Name, ".info") {
			segments := strings.Split(attrs.Name, "/")
			// version should be last segment w/ .info suffix
			last := segments[len(segments)-1]
			version := strings.TrimSuffix(last, ".info")
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		return nil, s.ErrNotFound{Module: module}
	}

	return versions, nil
}

func (b *gcpBucket) ObjectExists(ctx context.Context, module, version string) bool {
	_, err := b.Object(config.PackageVersionedName(module, version, "mod")).Attrs(ctx)
	return err == nil
}
