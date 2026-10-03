// Package toolchain implements the private storage layout for Go releases.
package toolchain

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
)

// Prefix is reserved for release objects, never module catalog entries.
const Prefix = ".athens/toolchains/v1/"

const maxArchiveSize = int64(1 << 30)

var filenameRE = regexp.MustCompile(`^go[0-9][A-Za-z0-9.\-]*\.(tar\.gz|zip|msi|pkg)$`)

// Objects is private backend plumbing. Create atomically publishes a complete
// object and returns KindAlreadyExists if the key already exists. Bodies passed
// to Create are seekable, with a known length.
type Objects interface {
	Open(ctx context.Context, key string) (storage.SizeReadCloser, error)
	Create(ctx context.Context, key string, body io.ReadSeeker, size int64) error
	Keys(ctx context.Context, prefix string) ([]string, error)
	Delete(ctx context.Context, key string) error
}

// Store implements the public release contracts using atomic backend objects.
type Store struct{ Objects Objects }

func sourcePath(source string) (string, error) {
	if source == "" {
		return "", errors.E("toolchain.source", "empty distribution identity", errors.KindBadRequest)
	}
	sum := sha256.Sum256([]byte(source))
	return Prefix + hex.EncodeToString(sum[:]) + "/", nil
}

func archivePath(source, filename string) (string, error) {
	prefix, err := sourcePath(source)
	if err != nil {
		return "", err
	}
	if !filenameRE.MatchString(filename) {
		return "", errors.E("toolchain.filename", "invalid release filename", errors.KindBadRequest)
	}
	return prefix + "archives/" + filename, nil
}

func (s *Store) Archive(ctx context.Context, source, filename string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	var info storage.ArchiveInfo
	key, err := archivePath(source, filename)
	if err != nil {
		return info, nil, err
	}
	body, err := s.Objects.Open(ctx, key)
	if err != nil {
		return info, nil, err
	}
	r := bufio.NewReaderSize(body, 16<<10)
	header, err := r.ReadSlice('\n')
	if err == nil {
		err = json.Unmarshal(header, &info)
	}
	if err != nil || info.Filename != filename || info.Size < 0 || body.Size()-int64(len(header)) != info.Size {
		_ = body.Close()
		return info, nil, errors.E("toolchain.Archive", "invalid stored archive metadata", errors.KindUnexpected)
	}
	return info, storage.NewSizer(&readerCloser{Reader: r, Closer: body}, info.Size), nil
}

type readerCloser struct {
	io.Reader
	io.Closer
}

func (s *Store) SaveArchive(ctx context.Context, source string, info storage.ArchiveInfo, body io.Reader) error {
	const op errors.Op = "toolchain.SaveArchive"
	key, err := archivePath(source, info.Filename)
	if err != nil {
		return err
	}
	if err := info.Validate(); err != nil {
		return err
	}
	digest, err := hex.DecodeString(info.SHA256)
	if err != nil || len(digest) != sha256.Size || info.Size <= 0 || info.Size > maxArchiveSize {
		return errors.E(op, "invalid archive size or SHA-256", errors.KindBadRequest)
	}
	// Temporary disk is staging only. The configured backend is authoritative.
	f, err := os.CreateTemp("", "athens-toolchain-*")
	if err != nil {
		return errors.E(op, err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if info.SavedAt.IsZero() {
		info.SavedAt = time.Now().UTC()
	}
	header, err := json.Marshal(info)
	if err != nil {
		return errors.E(op, err)
	}
	header = append(header, '\n')
	if _, err = f.Write(header); err != nil {
		return errors.E(op, err)
	}
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, hash), io.LimitReader(&contextReader{ctx, body}, info.Size+1))
	if err != nil {
		return errors.E(op, err)
	}
	if n != info.Size || !bytes.Equal(hash.Sum(nil), digest) {
		return errors.E(op, "archive size or SHA-256 mismatch", errors.KindBadRequest)
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return errors.E(op, err)
	}
	err = s.Objects.Create(ctx, key, f, int64(len(header))+n)
	if errors.Is(err, errors.KindAlreadyExists) {
		existing, rc, readErr := s.Archive(ctx, source, info.Filename)
		if readErr != nil {
			return readErr
		}
		_ = rc.Close()
		if strings.EqualFold(existing.SHA256, info.SHA256) && existing.Size == info.Size {
			return nil
		}
	}
	return err
}

type contextReader struct {
	//nolint:containedctx // io.Reader has no context parameter; cancellation follows each staged read.
	ctx  context.Context
	body io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.body.Read(p)
}

func (s *Store) ListArchives(ctx context.Context, source string) ([]storage.ArchiveInfo, error) {
	prefix, err := sourcePath(source)
	if err != nil {
		return nil, err
	}
	keys, err := s.Objects.Keys(ctx, prefix+"archives/")
	if err != nil {
		return nil, err
	}
	infos := make([]storage.ArchiveInfo, 0, len(keys))
	for _, key := range keys {
		info, body, err := s.Archive(ctx, source, path.Base(key))
		if errors.Is(err, errors.KindNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		_ = body.Close()
		infos = append(infos, info)
	}
	return infos, nil
}

func (s *Store) DeleteArchive(ctx context.Context, source, filename string) error {
	key, err := archivePath(source, filename)
	if err != nil {
		return err
	}
	return s.Objects.Delete(ctx, key)
}

func (s *Store) Releases(ctx context.Context, source string) (*storage.ReleaseList, error) {
	prefix, err := sourcePath(source)
	if err != nil {
		return nil, err
	}
	keys, err := s.Objects.Keys(ctx, prefix+"listings/")
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errors.E("toolchain.Releases", "release listing not found", errors.KindNotFound)
	}
	slices.Sort(keys)
	body, err := s.Objects.Open(ctx, keys[len(keys)-1])
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var list storage.ReleaseList
	if err := json.NewDecoder(io.LimitReader(body, 16<<20)).Decode(&list); err != nil {
		return nil, errors.E("toolchain.Releases", err)
	}
	return &list, nil
}

func (s *Store) SaveReleases(ctx context.Context, source string, list *storage.ReleaseList) error {
	prefix, err := sourcePath(source)
	if err != nil {
		return err
	}
	if list == nil || list.FetchedAt.IsZero() {
		return errors.E("toolchain.SaveReleases", "missing listing timestamp", errors.KindBadRequest)
	}
	body, err := json.Marshal(list)
	if err != nil {
		return err
	}
	if len(body) > 16<<20 {
		return errors.E("toolchain.SaveReleases", "listing exceeds size limit", errors.KindBadRequest)
	}
	// Immutable timestamped snapshots avoid out-of-order refresh overwrites.
	key := fmt.Sprintf("%slistings/%s.json", prefix, list.FetchedAt.UTC().Format("20060102T150405.000000000"))
	err = s.Objects.Create(ctx, key, bytes.NewReader(body), int64(len(body)))
	if errors.Is(err, errors.KindAlreadyExists) {
		return nil
	}
	if err != nil {
		return err
	}
	keys, err := s.Objects.Keys(ctx, prefix+"listings/")
	if err != nil {
		return err
	}
	slices.Sort(keys)
	// Retain two complete snapshots; an older concurrent writer cannot remove
	// a newer snapshot because all deletions precede this listing's timestamp.
	for i := 0; i < len(keys)-2; i++ {
		if keys[i] < key {
			if err := s.Objects.Delete(ctx, keys[i]); err != nil && !errors.Is(err, errors.KindNotFound) {
				return err
			}
		}
	}
	return nil
}
