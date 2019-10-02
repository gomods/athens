package paths

import (
	"net/http"
	"path"
	"strings"

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
	return DecodePath(version)
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

// MatchesPattern reports whether the path prefix of target matches
// pattern (as defined by path.Match)
//
// This tries to keep the same behavior as GOPRIVATE/GONOPROXY/GONOSUMDB,
// and is adopted from:
// https://github.com/golang/go/blob/a11644a26557ea436d456f005f39f4e01902bafe/src/cmd/go/internal/str/path.go#L58
func MatchesPattern(pattern, target string) bool {
	n := strings.Count(pattern, "/")
	prefix := target
	for i := 0; i < len(target); i++ {
		if target[i] == '/' {
			if n == 0 {
				prefix = target[:i]
				break
			}
			n--
		}
	}
	if n > 0 {
		return false
	}
	matched, _ := path.Match(pattern, prefix)
	return matched
}
