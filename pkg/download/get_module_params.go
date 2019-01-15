package download

import (
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
)

func getModuleParams(r *http.Request, op errors.Op) (mod string, ver string, err error) {
	params, err := paths.GetAllParams(r)
	if err != nil {
		return "", "", errors.E(op, err, errors.KindBadRequest)
	}

	return params.Module, params.Version, nil
}
