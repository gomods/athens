package download

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func (p *protocol) latestFromStorage(ctx context.Context, mod string) (*storage.RevInfo, error) {
	const op errors.Op = "protocol.latestFromStorage"
	versions, err := p.storage.List(ctx, mod)
	if err != nil {
		return nil, errors.E(op, err)
	}
	version := cachedLatestVersion(mod, versions)
	if version == "" {
		return nil, errors.E(op, errors.M(mod), "no cached version", errors.KindNotFound)
	}
	data, err := p.storage.Info(ctx, mod, version)
	if err != nil {
		return nil, errors.E(op, err)
	}
	var info storage.RevInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, errors.E(op, err)
	}
	if info.Version != version {
		return nil, errors.E(op, errors.M(mod), errors.V(version), "cached metadata does not match version")
	}
	return &info, nil
}

// cachedLatestVersion supplies the commit fallback for /@latest. The Go client
// has already considered tags from /@v/list, possibly excluding or retracting
// them, so returning a tag instead of an available commit can break resolution.
// Without a cached commit, tagged metadata is still useful to direct callers.
func cachedLatestVersion(mod string, versions []string) string {
	var pseudo, release, prerelease string
	var newest time.Time
	for _, version := range versions {
		if version != module.CanonicalVersion(version) || module.Check(mod, version) != nil {
			continue
		}
		switch {
		case module.IsPseudoVersion(version):
			stamp, err := module.PseudoVersionTime(version)
			if err != nil {
				continue
			}
			if stamp.After(newest) || (stamp.Equal(newest) && semver.Compare(version, pseudo) > 0) {
				pseudo, newest = version, stamp
			}
		case semver.Prerelease(version) == "":
			if semver.Compare(version, release) > 0 {
				release = version
			}
		default:
			if semver.Compare(version, prerelease) > 0 {
				prerelease = version
			}
		}
	}
	if pseudo != "" {
		return pseudo
	}
	if release != "" {
		return release
	}
	return prerelease
}
