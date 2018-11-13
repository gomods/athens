package paths

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gorilla/mux"
)

// GetModule gets the module from the path of a ?go-get=1 request
func GetModule(c buffalo.Context) (string, error) {
	const op errors.Op = "paths.GetModule"

	module := c.Param("module")
	if module == "" {
		return "", errors.E(op, "missing module parameter")
	}

	return DecodePath(module)
}

// GetModuleFromRequest gets a module from the path of an http.Request
func GetModuleFromRequest(r *http.Request) (string, error) {
	const op errors.Op = "paths.GetModule"

	module, ok := mux.Vars(r)["module"]
	if module == "" || !ok {
		return "", errors.E(op, "missing module parameter")
	}

	return DecodePath(module)
}

// GetVersion gets the version from the path of a ?go-get=1 request
func GetVersion(c buffalo.Context) (string, error) {
	const op errors.Op = "paths.GetVersion"

	version := c.Param("version")
	if version == "" {
		return "", errors.E(op, "missing version paramater")
	}
	return version, nil
}

// AllPathParams holds the module and version in the path of a ?go-get=1
// request
type AllPathParams struct {
	Module  string
	Version string
}

// GetAllParams fetches the path patams from c and returns them
func GetAllParams(c buffalo.Context) (*AllPathParams, error) {
	const op errors.Op = "paths.GetAllParams"
	mod, err := GetModule(c)
	if err != nil {
		return nil, errors.E(op, err)
	}

	version, err := GetVersion(c)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &AllPathParams{Module: mod, Version: version}, nil
}
