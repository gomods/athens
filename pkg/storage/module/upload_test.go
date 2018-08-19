package module

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

type UploadTests struct {
	suite.Suite
	timeout time.Duration
}

func TestUpload(t *testing.T) {
	conf := getConf(t)
	timeout := config.TimeoutDuration(conf.Timeout)
	suite.Run(t, &UploadTests{
		timeout: timeout,
	})
}

func (u *UploadTests) TestUploadTimeout() {
	r := u.Require()
	rd := bytes.NewReader([]byte("123"))
	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithTimeout, u.timeout)
	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "uploading mx.1.1.1.info failed: context deadline exceeded")
	r.Contains(err.Error(), "uploading mx.1.1.1.zip failed: context deadline exceeded")
	r.Contains(err.Error(), "uploading mx.1.1.1.mod failed: context deadline exceeded")
}

func (u *UploadTests) TestUploadError() {
	r := u.Require()
	rd := bytes.NewReader([]byte("123"))
	err := Upload(context.Background(), "mx", "1.1.1", rd, rd, rd, uplWithErr, u.timeout)
	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "some err")
}

func uplWithTimeout(ctx context.Context, path, contentType string, stream io.Reader) error {
	time.Sleep(2 * time.Second)
	return nil
}

func uplWithErr(ctx context.Context, path, contentType string, stream io.Reader) error {
	return errors.New("some err")
}
