package paths

import (
	"net/http"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gorilla/mux"
)

// GetModule gets the module from the path of a ?go-get=1 request
func GetModule(r *http.Request) (string, error) {
	const op errors.Op = "paths.GetModule"
	module := mux.Vars(r)["module"]
	if module == "" {
		return "", errors.E(op, "missing module parameter")
	}
	return DecodePath(module)
}

// GetVersion gets the version from the path of a ?go-get=1 request
func GetVersion(r *http.Request) (string, error) {
	const op errors.Op = "paths.GetVersion"

	version := mux.Vars(r)["version"]
	if version == "" {
		return "", errors.E(op, "missing version parameter")
	}
	return version, nil
}

// AllPathParams holds the module and version in the path of a ?go-get=1
// request
type AllPathParams struct {
	Module  string `json:"module"`
	Version string `json:"version"`
}

// GetAllParams fetches the path params from r and returns them
func GetAllParams(r *http.Request) (*AllPathParams, error) {
	const op errors.Op = "paths.GetAllParams"
	mod, err := GetModule(r)
	if err != nil {
		return nil, errors.E(op, err)
	}

	version, err := GetVersion(r)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &AllPathParams{Module: mod, Version: version}, nil
}
