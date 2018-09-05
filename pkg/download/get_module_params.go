package download

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
)

func getModuleParams(op errors.Op, c buffalo.Context) (mod string, vers string, err error) {
	params, err := paths.GetAllParams(c)

	return params.Module, params.Version, errors.E(op, err, errors.KindBadRequest)
}
