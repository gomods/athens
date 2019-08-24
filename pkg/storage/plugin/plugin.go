package plugin

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stpb "github.com/gomods/athens/pkg/storage/plugin/pb/v1/storage"
)

// const msgLimit = 64 * 1024    // 64 KiB
// const bufferSize = 256 * 1024 // 256 KiB

type backend struct {
	b storage.Backend
}

func (sb backend) List(ctx context.Context, req *stpb.ListRequest) (*stpb.ListResponse, error) {
	ls, err := sb.b.List(ctx, req.GetModule())
	if err != nil {
		return nil, err
	}
	return &stpb.ListResponse{Versions: ls}, nil
}

func (sb backend) GetInfo(ctx context.Context, req *stpb.GetModuleRequest) (*stpb.GetModuleResponse, error) {
	inf, err := sb.b.Info(ctx, req.GetModule(), req.GetVersion())
	if err != nil {
		return nil, err
	}
	return &stpb.GetModuleResponse{Data: inf}, nil
}

func (sb backend) GetMod(ctx context.Context, req *stpb.GetModuleRequest) (*stpb.GetModuleResponse, error) {
	mod, err := sb.b.GoMod(ctx, req.GetModule(), req.GetVersion())
	if err != nil {
		return nil, err
	}
	return &stpb.GetModuleResponse{Data: mod}, nil
}

func (sb backend) GetZip(req *stpb.GetModuleRequest, str stpb.StorageBackendService_GetZipServer) error {
	zip, err := sb.b.Zip(str.Context(), req.GetModule(), req.GetVersion())
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	buf.Grow(msgLimit)
	msg := &stpb.GetModuleResponse{}
	for err == nil {
		buf.Reset()
		_, err := io.CopyN(buf, zip, msgLimit)
		if err != nil && err != io.EOF {
			fmt.Println("copy error", err)
			return err
		}
		msg.Data = buf.Bytes()
		if err := str.Send(msg); err != nil {
			fmt.Println("send error", err)
			return err
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func (sb backend) Exists(ctx context.Context, req *stpb.ExistsRequest) (*stpb.ExistsResponse, error) {
	exist, err := sb.b.Exists(ctx, req.GetModule(), req.GetVersion())
	if err != nil {
		return nil, err
	}
	return &stpb.ExistsResponse{Exists: exist}, nil
}

func (sb backend) Save(str stpb.StorageBackendService_SaveServer) error {
	msg, err := str.Recv()
	if err != nil {
		return err
	}
	mod := msg.GetModDefinition()
	if mod == nil {
		return status.Error(codes.InvalidArgument, "missing module definition")
	}
	r, w := io.Pipe()
	defer r.Close()
	go func() {
		b := bufio.NewWriterSize(w, bufferSize)
		_, err := b.Write(msg.GetZip())
		for err == nil {
			msg, err = str.Recv()
			if err != nil {
				break
			}
			_, err = b.Write(msg.GetZip())
		}
		if err == io.EOF {
			err = b.Flush()
		}
		if err != nil {
			w.CloseWithError(status.Error(codes.NotFound, err.Error()))
		}
		w.Close()
	}()
	err = sb.b.Save(str.Context(), mod.GetModule(), mod.GetVersion(), mod.GetMod(), r, mod.GetInfo())
	if err != nil {
		return err
	}
	return str.SendAndClose(&stpb.SaveResponse{})
}

func (sb backend) Delete(ctx context.Context, req *stpb.DeleteRequest) (*stpb.DeleteResponse, error) {
	err := sb.b.Delete(ctx, req.GetModule(), req.GetVersion())
	if err != nil {
		return nil, err
	}
	return &stpb.DeleteResponse{}, nil
}
