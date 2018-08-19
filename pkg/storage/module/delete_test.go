package module

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/stretchr/testify/suite"
)

type DeleteTests struct {
	suite.Suite
	timeout time.Duration
}

const (
	testConfigFile = "../../../config.test.toml"
)

func getConf(t *testing.T) *config.Config {
	fmt.Printf("Config File: %s", testConfigFile)
	absPath, err := filepath.Abs(testConfigFile)
	if err != nil {
		t.Errorf("Unable to construct absolute path to test config file")
	}
	conf, err := config.ParseConfigFile(absPath)
	if err != nil {
		t.Errorf("Unable to parse config file")
	}
	return conf
}

func TestDelete(t *testing.T) {
	conf := getConf(t)
	timeout := config.TimeoutDuration(conf.Timeout)
	fmt.Printf("TIMEOUT: %d\n", timeout/time.Second)
	suite.Run(t, &DeleteTests{
		timeout: timeout,
	})
}

func (d *DeleteTests) TestDeleteTimeout() {
	r := d.Require()

	err := Delete(context.Background(), "mx", "1.1.1", delWithTimeout, d.timeout)

	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "deleting mx.1.1.1.info failed: context deadline exceeded")
	r.Contains(err.Error(), "deleting mx.1.1.1.zip failed: context deadline exceeded")
	r.Contains(err.Error(), "deleting mx.1.1.1.mod failed: context deadline exceeded")
}

func (d *DeleteTests) TestDeleteError() {
	r := d.Require()

	err := Delete(context.Background(), "mx", "1.1.1", delWithErr, d.timeout)

	r.Error(err, "deleter returned at least one error")
	r.Contains(err.Error(), "some err")
}

func delWithTimeout(ctx context.Context, path string) error {
	time.Sleep(2 * time.Second)
	return nil
}

func delWithErr(ctx context.Context, path string) error {
	return errors.New("some err")
}
