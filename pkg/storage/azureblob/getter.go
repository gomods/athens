package azureblob

import (
	"context"
	"io"
)

func (s *Storage) Info(ctx context.Context, module string, vsn string) ([]byte, error) {
	panic("not implemented")
}

func (s *Storage) GoMod(ctx context.Context, module string, vsn string) ([]byte, error) {
	panic("not implemented")
}

func (s *Storage) Zip(ctx context.Context, module string, vsn string) (io.ReadCloser, error) {
	panic("not implemented")
}
