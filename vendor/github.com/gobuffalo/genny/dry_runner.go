package genny

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DryRunner will NOT execute commands and write files
// it is NOT destructive
func DryRunner(ctx context.Context) *Runner {
	pwd, _ := os.Getwd()
	l := logrus.New()
	l.Out = os.Stdout
	l.SetLevel(logrus.DebugLevel)
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
