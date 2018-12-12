package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PackageVersionedName return package full name used in storage.
// E.g athens/@v/v1.0.mod
func PackageVersionedName(module, version, ext string) string {
	return fmt.Sprintf("%s/@v/%s.%s", module, version, ext)
}

// FmtModVer is a helper function that can take
// pkg/a/b and v2.3.1 and returns pkg/a/b@v2.3.1
func FmtModVer(mod, ver string) string {
	return fmt.Sprintf("%s@%s", mod, ver)
}

// ModuleVersionFromPath returns module and version from a
// storage path
// E.g athens/@v/v1.0.info -> athens and v.1.0
func ModuleVersionFromPath(path string) (string, string) {
	segments := strings.Split(path, "/@v/")
	if len(segments) != 2 {
		return "", ""
	}
	version := strings.TrimSuffix(segments[1], filepath.Ext(segments[1]))
	return segments[0], version
}
