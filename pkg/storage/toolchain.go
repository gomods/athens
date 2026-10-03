package storage

import (
	"context"
	"encoding/hex"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/errors"
)

// Release describes a Go release in the go.dev/dl JSON format.
type Release struct {
	Version string        `json:"version"`
	Stable  bool          `json:"stable"`
	Files   []ReleaseFile `json:"files"`
}

// ReleaseFile identifies an archive or installer and its upstream checksum.
type ReleaseFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

// ArchiveInfo describes a verified archive saved by Athens.
type ArchiveInfo struct {
	ReleaseFile

	Stable  bool      `json:"stable"`
	SavedAt time.Time `json:"saved_at"`
}

// ReleaseList is a persisted upstream listing and the time it was fetched.
type ReleaseList struct {
	Releases  []Release `json:"releases"`
	FetchedAt time.Time `json:"fetched_at"`
}

// ArchiveReader reads only complete, verified archives from a distribution.
// Source is a stable distribution identity, not a physical storage path.
type ArchiveReader interface {
	Archive(ctx context.Context, source, filename string) (ArchiveInfo, SizeReadCloser, error)
	ListArchives(ctx context.Context, source string) ([]ArchiveInfo, error)
}

// ArchiveWriter publishes and deletes archives. SaveArchive checks size and
// SHA-256 before publication. An identical save is idempotent; different bytes
// for an existing filename return KindAlreadyExists. Failed saves stay invisible.
type ArchiveWriter interface {
	SaveArchive(ctx context.Context, source string, info ArchiveInfo, body io.Reader) error
	DeleteArchive(ctx context.Context, source, filename string) error
}

// ReleaseStorage persists discovery metadata independently of archive inventory.
// Saving an older listing must not replace a newer one.
type ReleaseStorage interface {
	Releases(ctx context.Context, source string) (*ReleaseList, error)
	SaveReleases(ctx context.Context, source string, list *ReleaseList) error
}

// ToolchainStorage is an optional capability of a module storage backend.
// Release archives are kept outside the module catalog and checksum database.
type ToolchainStorage interface {
	ArchiveReader
	ArchiveWriter
	ReleaseStorage
}

var releaseVersionRE = regexp.MustCompile(`^go[0-9]+(?:\.[0-9]+){0,2}(?:(?:beta|rc)[0-9]+)?$`)

// Validate checks that a file's identity, size and checksum describe the same
// Go release archive. Both upstream discovery and storage publication use it.
func (f ReleaseFile) Validate() error {
	digest, err := hex.DecodeString(f.SHA256)
	if err != nil || len(digest) != 32 || f.Size <= 0 || f.Size > 1<<30 || !releaseVersionRE.MatchString(f.Version) {
		return errors.E("storage.ReleaseFile", "invalid release size, checksum or version", errors.KindBadRequest)
	}
	prefix := f.Version + "." + f.OS + "-" + f.Arch + "."
	if f.Kind == "source" {
		prefix = f.Version + ".src."
		if f.OS != "" || f.Arch != "" {
			return errors.E("storage.ReleaseFile", "source archive must not specify a platform", errors.KindBadRequest)
		}
	} else if f.Kind != "archive" && f.Kind != "installer" {
		return errors.E("storage.ReleaseFile", "invalid release file kind", errors.KindBadRequest)
	}
	// Older Go distributions label ARM metadata armv6l while some archive
	// names use arm. Keep the explicit alias instead of weakening identity checks.
	if f.Arch == "armv6l" && strings.HasPrefix(f.Filename, f.Version+"."+f.OS+"-arm.") {
		prefix = f.Version + "." + f.OS + "-arm."
	}
	// go.dev also publishes the two dated Go 1.4 bootstrap archives under
	// version go1, OS 4 and architecture bootstrap.
	if f.Version == "go1" && f.OS == "4" && f.Arch == "bootstrap" && f.Kind == "archive" &&
		(f.Filename == "go1.4-bootstrap-20170518.tar.gz" || f.Filename == "go1.4-bootstrap-20170531.tar.gz") {
		return nil
	}
	if !strings.HasPrefix(f.Filename, prefix) {
		return errors.E("storage.ReleaseFile", "filename does not match release version and platform", errors.KindBadRequest)
	}
	extension := strings.TrimPrefix(f.Filename, prefix)
	if (f.Kind == "archive" && (extension == "tar.gz" || extension == "zip")) || (f.Kind == "source" && extension == "tar.gz") || (f.Kind == "installer" && (extension == "pkg" || extension == "msi")) {
		return nil
	}
	return errors.E("storage.ReleaseFile", "invalid release archive extension", errors.KindBadRequest)
}
