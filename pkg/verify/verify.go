package verify

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gomods/athens/pkg/storage"
	"golang.org/x/mod/sumdb/dirhash"
)

// Poisoned reports whether the stored zip for mod@ver disagrees with the
// canonical hash from the oracle. Any error obtaining the stored zip or the
// canonical hash is returned to the caller, which MUST treat it as
// "cannot verify -> keep" (never delete on uncertainty).
func Poisoned(ctx context.Context, g storage.Getter, o Oracle, mod, ver string) (bool, error) {
	rc, err := g.Zip(ctx, mod, ver)
	if err != nil {
		return false, fmt.Errorf("verify: read stored zip %s@%s: %w", mod, ver, err)
	}
	defer rc.Close()

	stored, err := hashZipReader(rc)
	if err != nil {
		return false, fmt.Errorf("verify: hash stored zip %s@%s: %w", mod, ver, err)
	}
	canonical, err := o.CanonicalHash(ctx, mod, ver)
	if err != nil {
		return false, err
	}
	return stored != canonical, nil
}

// hashZipReader spools a zip to a temp file (dirhash.HashZip needs a path) and
// returns its h1: hash.
func hashZipReader(r io.Reader) (string, error) {
	f, err := os.CreateTemp("", "athens-verify-*.zip")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return dirhash.HashZip(f.Name(), dirhash.Hash1)
}
