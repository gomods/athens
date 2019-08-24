package plugin

import (
	"bytes"
	"context"
	"io"

	"github.com/gomods/athens/pkg/storage"

	stpb "github.com/gomods/athens/pkg/storage/plugin/pb/v1/storage"
)

const msgLimit = 64 * 1024    // 64 KiB
const bufferSize = 256 * 1024 // 256 KiB

var _ storage.Backend = LocalConn{}

// List the stored versions of a given module.
func (p LocalConn) List(ctx context.Context, module string) ([]string, error) {
	resp, err := p.c.List(ctx, &stpb.ListRequest{Module: module})
	if err != nil {
		return nil, err
	}
	return resp.GetVersions(), nil
}

// Info file from storage
func (p LocalConn) Info(ctx context.Context, module string, vsn string) ([]byte, error) {
	resp, err := p.c.GetInfo(ctx, &stpb.GetModuleRequest{Module: module, Version: vsn})
	if err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

// GoMod get the go.mod file from storage.
func (p LocalConn) GoMod(ctx context.Context, module string, vsn string) ([]byte, error) {
	resp, err := p.c.GetMod(ctx, &stpb.GetModuleRequest{Module: module, Version: vsn})
	if err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

var _ io.ReadCloser = zipReader{}

type zipReader struct {
	b *bytes.Buffer
	s stpb.StorageBackendService_GetZipClient
}

// Read from the zip stream.
func (z zipReader) Read(b []byte) (int, error) {
	if z.b.Len() >= len(b) || z.b.Len() >= bufferSize {
		return z.b.Read(b)
	}
	msg, err := z.s.Recv()
	if err != nil {
		if err == io.EOF {
			return z.b.Read(b)
		}
		return 0, err
	}
	if _, err := z.b.Write(msg.GetData()); err != nil {
		return 0, err
	}
	return z.b.Read(b)
}

// Close the underlying io.
func (z zipReader) Close() error {
	z.b.Reset()
	return z.s.CloseSend()
}

// Zip fetches zip file from storage.
func (p LocalConn) Zip(ctx context.Context, module string, vsn string) (io.ReadCloser, error) {
	str, err := p.c.GetZip(ctx, &stpb.GetModuleRequest{Module: module, Version: vsn})
	if err != nil {
		return nil, err
	}
	msg, err := str.Recv()
	if err != nil {
		return nil, err
	}
	return zipReader{
		b: bytes.NewBuffer(msg.GetData()),
		s: str,
	}, nil
}

// Exists check for the module version in storage.
func (p LocalConn) Exists(ctx context.Context, module string, version string) (bool, error) {
	resp, err := p.c.Exists(ctx, &stpb.ExistsRequest{Module: module, Version: version})
	if err != nil {
		return false, err
	}
	return resp.GetExists(), nil
}

// Save version information for a module version
func (p LocalConn) Save(ctx context.Context, module string, version string, mod []byte, zip io.Reader, info []byte) error {
	str, err := p.c.Save(ctx)
	if err != nil {
		return err
	}
	err = str.Send(&stpb.SaveRequest{
		ModDefinition: &stpb.SaveRequest_Module{
			Module:  module,
			Version: version,
			Mod:     mod,
			Info:    info,
		},
	})
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	msg := &stpb.SaveRequest{}
	for err == nil {
		buf.Reset()
		_, err := io.CopyN(buf, zip, msgLimit)
		if err != nil && err != io.EOF {
			return err
		}
		msg.Zip = buf.Bytes()
		if err := str.Send(msg); err != nil {
			return err
		}
		if err == io.EOF {
			break
		}
	}
	_, err = str.CloseAndRecv()
	return err
}

// Delete a module version from storage.
func (p LocalConn) Delete(ctx context.Context, module string, vsn string) error {
	_, err := p.c.Delete(ctx, &stpb.DeleteRequest{Module: module, Version: vsn})
	return err
}
