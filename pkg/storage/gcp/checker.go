package gcp

import (
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists implements the (./pkg/storage).Checker interface
// returning true if the module at version exists in storage
func (s *Storage) Exists(ctx observ.ProxyContext, module, version string) (bool, error) {
	const op errors.Op = "gcp.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	return s.bucket.Exists(ctx, config.PackageVersionedName(module, version, "mod"))
}
