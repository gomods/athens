package http

import (
	"context"
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
)

// Exists checks for a specific version of a module
func (s *ModuleStore) Exists(ctx context.Context, module, vsn string) (bool, error) {
	var op errors.Op = "http.Exists"
	ctx, span := observ.StartSpan(ctx, op.String())
	defer span.End()

	req, _ := http.NewRequest(http.MethodHead, s.moduleRoot(module)+vsn+".mod", nil)
	req.SetBasicAuth(s.username, s.password)
	resp, err := s.client.Do(req)
	if err != nil {
		return false, errors.E(op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		kind := errors.KindUnexpected
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, errors.E(op, kind, errors.M(module), errors.V(vsn))
	}

	return true, nil
}
