package module

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/suite"
)

type UploadTests struct {
	suite.Suite
}

func TestUpload(t *testing.T) {
	suite.Run(t, new(UploadTests))
}

func (m *UploadTests) TestUploadTimeout() {
	r := m.Require()
	envy.Set("ATHENS_TIMEOUT", "1")
	defer envy.Set("ATHENS_TIMEOUT", "300")
	rd := bytes.NewReader([]byte("123"))
	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithTimeout)

	me := err.(*multierror.Error)
	r.Equal(3, len(me.WrappedErrors()))
	r.Contains(me.Error(), "uploading mx.1.1.1.info failed: context deadline exceeded")
	r.Contains(me.Error(), "uploading mx.1.1.1.zip failed: context deadline exceeded")
	r.Contains(me.Error(), "uploading mx.1.1.1.mod failed: context deadline exceeded")
}

func (m *UploadTests) TestUploadError() {
	r := m.Require()
	envy.Set("ATHENS_TIMEOUT", "1")
	defer envy.Set("ATHENS_TIMEOUT", "300")
	rd := bytes.NewReader([]byte("123"))
	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithErr)

	me := err.(*multierror.Error)
	r.Equal(3, len(me.WrappedErrors()))
	r.Contains(me.Error(), "some err")
	r.Contains(me.Error(), "some err")
	r.Contains(me.Error(), "some err")
}

func uplWithTimeout(ctx context.Context, path, contentType string, stream io.Reader) error {
	time.Sleep(2 * time.Second)
	return nil
}

func uplWithErr(ctx context.Context, path, contentType string, stream io.Reader) error {
	return errors.New("some err")
}
