package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/stretchr/testify/suite"
)

type UploadTests struct {
	suite.Suite
}

func TestUpload(t *testing.T) {
	suite.Run(t, new(UploadTests))
}

func (u *UploadTests) SetupTest() {
	envy.Set("ATHENS_TIMEOUT", "1")
}

func (u *UploadTests) TearDownTest() {
	envy.Set("ATHENS_TIMEOUT", "300")
}

func (u *UploadTests) TestUploadTimeout() {
	r := u.Require()
	rd := NewStreamFromBytes([]byte("123"))
	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithTimeout, time.Second)
	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "uploading mx.1.1.1.info failed: context deadline exceeded")
	r.Contains(err.Error(), "uploading mx.1.1.1.zip failed: context deadline exceeded")
	r.Contains(err.Error(), "uploading mx.1.1.1.mod failed: context deadline exceeded")
}

func (u *UploadTests) TestUploadError() {
	r := u.Require()
	rd := NewStreamFromBytes([]byte("123"))

	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithErr, time.Second)
	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "some err")
}

func uplWithTimeout(ctx context.Context, path, contentType string, stream Stream) error {
	time.Sleep(2 * time.Second)
	return nil
}

func uplWithErr(ctx context.Context, path, contentType string, stream Stream) error {
	return errors.New("some err")
}
