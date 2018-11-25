package module

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/observ"
	"github.com/stretchr/testify/suite"
)

type DeleteTests struct {
	suite.Suite
}

func TestDelete(t *testing.T) {
	suite.Run(t, new(DeleteTests))
}

func (d *DeleteTests) SetupTest() {
	envy.Set("ATHENS_TIMEOUT", "1")
}

func (d *DeleteTests) TearDownTest() {
	envy.Set("ATHENS_TIMEOUT", "300")
}

func (d *DeleteTests) TestDeleteTimeout() {
	r := d.Require()

	err := Delete(context.Background(), "mx", "1.1.1", delWithTimeout, time.Second)

	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "deleting mx.1.1.1.info failed: context deadline exceeded")
	r.Contains(err.Error(), "deleting mx.1.1.1.zip failed: context deadline exceeded")
	r.Contains(err.Error(), "deleting mx.1.1.1.mod failed: context deadline exceeded")
}

func (d *DeleteTests) TestDeleteError() {
	r := d.Require()

	err := Delete(context.Background(), "mx", "1.1.1", delWithErr, time.Second)

	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "some err")
}

func delWithTimeout(ctx observ.ProxyContext, path string) error {
	time.Sleep(2 * time.Second)
	return nil
}

func delWithErr(ctx observ.ProxyContext, path string) error {
	return errors.New("some err")
}
