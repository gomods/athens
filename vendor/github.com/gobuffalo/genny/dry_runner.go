package genny

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"

	"github.com/gobuffalo/logger"
	"github.com/pkg/errors"
)

// DryRunner will NOT execute commands and write files
// it is NOT destructive
func DryRunner(ctx context.Context) *Runner {
	pwd, _ := os.Getwd()
	l := logger.New(logger.DebugLevel)
	r := &Runner{
		Logger:  l,
		Context: ctx,
		Root:    pwd,
		moot:    &sync.RWMutex{},
		FileFn: func(f File) (File, error) {
			bb := &bytes.Buffer{}
			mw := io.MultiWriter(bb, os.Stdout)
			if _, err := io.Copy(mw, f); err != nil {
				return f, errors.WithStack(err)
			}
			return NewFile(f.Name(), bb), nil
		},
	}
	r.Disk = newDisk(r)
	return r
}
