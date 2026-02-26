package fs

import (
		"context"
		"fmt"
		"os"

		"github.com/gomods/athens/pkg/errors"
		"github.com/gomods/athens/pkg/observ"
		"github.com/spf13/afero"
	)

// expectedFiles returns the three file names that must exist for a module version.
func expectedFiles(version string) []string {
		return []string{
					version + ".info",
					version + ".mod",
					version + ".zip",
				}
}

func (s *storageImpl) Exists(ctx context.Context, module, version string) (bool, error) {
		const op errors.Op = "fs.Exists"
		_, span := observ.StartSpan(ctx, op.String())
		defer span.End()
		versionedPath := s.versionLocation(module, version)

		files, err := afero.ReadDir(s.filesystem, versionedPath)
		if err != nil {
					if os.IsNotExist(err) {
									return false, nil
								}
					return false, errors.E(op, errors.M(module), errors.V(version), err)
				}

		if len(files) == 3 {
					return true, nil
				}

		// Identify which of the three required files are missing and warn.
		present := make(map[string]bool, len(files))
		for _, f := range files {
					present[f.Name()] = true
				}
		var missing []string
		for _, name := range expectedFiles(version) {
					if !present[name] {
									missing = append(missing, name)
								}
				}
		if len(missing) > 0 {
					span.AddEvent(fmt.Sprintf(
									"incomplete module storage: %s@%s is missing files: %v — falling back to VCS",
									module, version, missing,
								))
				}

		return false, nil
}
