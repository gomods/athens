//go:build !unix

package shutdown

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// GetSignals returns the appropriate signals to catch for a clean shutdown, dependent on the OS.
func GetSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

// ChildProcReaper spawns a goroutine to listen for SIGCHLD signals to cleanup
// zombie child processes. The returned context will be canceled once all child
// processes have been cleaned up, and should be waited on before exiting.
//
// This only applies to Unix platforms, and returns an already canceled context
// on Windows.
func ChildProcReaper(ctx context.Context, logger logrus.FieldLogger) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return ctx
}
