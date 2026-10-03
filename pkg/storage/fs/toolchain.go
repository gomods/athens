package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/internal/toolchain"
	"github.com/spf13/afero"
)

func (s *storageImpl) toolchains() *toolchain.Store {
	return &toolchain.Store{Objects: &releaseObjects{s}}
}

func (s *storageImpl) Archive(ctx context.Context, source, filename string) (storage.ArchiveInfo, storage.SizeReadCloser, error) {
	return s.toolchains().Archive(ctx, source, filename)
}

func (s *storageImpl) SaveArchive(ctx context.Context, source string, info storage.ArchiveInfo, body io.Reader) error {
	return s.toolchains().SaveArchive(ctx, source, info, body)
}

func (s *storageImpl) ListArchives(ctx context.Context, source string) ([]storage.ArchiveInfo, error) {
	return s.toolchains().ListArchives(ctx, source)
}

func (s *storageImpl) DeleteArchive(ctx context.Context, source, filename string) error {
	return s.toolchains().DeleteArchive(ctx, source, filename)
}

func (s *storageImpl) Releases(ctx context.Context, source string) (*storage.ReleaseList, error) {
	return s.toolchains().Releases(ctx, source)
}

func (s *storageImpl) SaveReleases(ctx context.Context, source string, list *storage.ReleaseList) error {
	return s.toolchains().SaveReleases(ctx, source, list)
}

type releaseObjects struct{ store *storageImpl }

func (o *releaseObjects) Open(ctx context.Context, key string) (storage.SizeReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f, err := o.store.filesystem.Open(filepath.Join(o.store.rootDir, key, "data"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.E("fs.Archive", err, errors.KindNotFound)
		}
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return storage.NewSizer(f, info.Size()), nil
}

func (o *releaseObjects) Create(ctx context.Context, key string, body io.ReadSeeker, size int64) error {
	o.store.archiveMu.Lock()
	defer o.store.archiveMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	final := filepath.Join(o.store.rootDir, key)
	exists, err := afero.Exists(o.store.filesystem, filepath.Join(final, "data"))
	if err != nil {
		return err
	}
	if exists {
		return errors.E("fs.SaveArchive", "already stored", errors.KindAlreadyExists)
	}
	// Reclaim only an empty directory left by an interrupted legacy delete.
	if err := o.store.filesystem.Remove(final); err != nil && !os.IsNotExist(err) {
		if exists, _ := afero.Exists(o.store.filesystem, filepath.Join(final, "data")); exists {
			return errors.E("fs.SaveArchive", "already stored", errors.KindAlreadyExists)
		}
		return err
	}
	parent := filepath.Dir(final)
	if err := o.store.filesystem.MkdirAll(parent, 0o750); err != nil {
		return err
	}
	stage, err := afero.TempDir(o.store.filesystem, parent, ".staging-")
	if err != nil {
		return err
	}
	defer o.store.filesystem.RemoveAll(stage)
	f, err := o.store.filesystem.OpenFile(filepath.Join(stage, "data"), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, body)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if n != size {
		return errors.E("fs.SaveArchive", "incomplete object")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Publishing a nonempty directory is atomic and cannot replace another
	// published nonempty directory, including across Athens processes.
	if err := o.store.filesystem.Rename(stage, final); err != nil {
		exists, checkErr := afero.Exists(o.store.filesystem, final)
		if checkErr == nil && exists {
			return errors.E("fs.SaveArchive", "already stored", errors.KindAlreadyExists)
		}
		return err
	}
	return nil
}

func (o *releaseObjects) Keys(ctx context.Context, prefix string) ([]string, error) {
	root := filepath.Join(o.store.rootDir, prefix)
	exists, err := afero.Exists(o.store.filesystem, root)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	var keys []string
	err = afero.Walk(o.store.filesystem, root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".staging-") {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() == "data" {
			key, err := filepath.Rel(o.store.rootDir, filepath.Dir(p))
			if err != nil {
				return err
			}
			keys = append(keys, filepath.ToSlash(key))
		}
		return nil
	})
	return keys, err
}

func (o *releaseObjects) Delete(ctx context.Context, key string) error {
	o.store.archiveMu.Lock()
	defer o.store.archiveMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	p := filepath.Join(o.store.rootDir, key)
	exists, err := afero.Exists(o.store.filesystem, p)
	if err != nil {
		return err
	}
	if !exists {
		return errors.E("fs.DeleteArchive", "archive not found", errors.KindNotFound)
	}
	tombstone, err := afero.TempDir(o.store.filesystem, filepath.Dir(p), ".staging-delete-")
	if err != nil {
		return err
	}
	defer o.store.filesystem.RemoveAll(tombstone)
	if err := o.store.filesystem.Remove(tombstone); err != nil {
		return err
	}
	// Removing the public name is atomic; cleanup can never leave an empty
	// published directory that blocks the next save.
	if err := o.store.filesystem.Rename(p, tombstone); err != nil {
		return err
	}
	return o.store.filesystem.RemoveAll(tombstone)
}
