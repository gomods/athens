package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/observ"
	"github.com/spf13/afero"
)

func (s *storageImpl) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	const op errors.Op = "fs.Save"
	_, span := observ.StartSpan(ctx, op.String())
	defer span.End()
	dir := s.versionLocation(module, version)

	// NB: The process's umask is subtracted from the permissions below,
	// so an umask of for example 0077 allows directories and files to be
	// created with mode 0700 / 0600, i.e. not world- or group-readable.

	// Make the versioned directory to hold the go.mod and the zipfile.
	if err := s.filesystem.MkdirAll(dir, 0o777); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	// Write the go.mod file.
	if err := afero.WriteFile(s.filesystem, filepath.Join(dir, "go.mod"), mod, 0o666); err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	// Write the zipfile.
	f, err := s.filesystem.OpenFile(filepath.Join(dir, "source.zip"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, zip)
	if err != nil {
		return errors.E(op, err, errors.M(module), errors.V(version))
	}

	// write the info file
	err = afero.WriteFile(s.filesystem, filepath.Join(dir, version+".info"), info, 0o666)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}
