package gcp

import (
	"bytes"
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
	stg "github.com/gomods/athens/pkg/storage"
	m "github.com/gomods/athens/pkg/storage/module"
)

// Save uploads the module's .mod, .zip and .info files for a given version
// It expects a context, which can be provided using context.Background
// from the standard library until context has been threaded down the stack.
// see issue: https://github.com/gomods/athens/issues/174
//
// Uploaded files are publicly accessable in the storage bucket as per
// an ACL rule.
func (s *Storage) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	if exists := s.Exists(module, version); exists {
		return stg.ErrVersionAlreadyExists{Module: module, Version: version}
	}

	err := m.Upload(ctx, module, version, bytes.NewReader(info), bytes.NewReader(mod), zip, s.upload)
	// TODO: take out lease on the /list file and add the version to it
	//
	// Do that only after module source+metadata is uploaded
	return err
}

func (s *Storage) upload(ctx context.Context, path, contentType string, stream io.Reader) error {
	wc := s.bucket.Object(path).NewWriter(ctx)
	defer func(w *storage.Writer) {
		if err := w.Close(); err != nil {
			log.Printf("WARNING: failed to close storage object writer: %s", err)
		}
	}(wc)
	wc.ContentType = contentType
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err := io.Copy(wc, stream); err != nil {
		return err
	}
	return nil
}
