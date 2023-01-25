//go:build unix

package shutdown

import (
	"os"
	"syscall"
)

// GetSignals returns the appropriate signals to catch for a clean shutdown, dependent on the OS.
//
// On Unix-like operating systems, it is important to catch SIGTERM in addition to SIGINT.
func GetSignals() []os.Signal {
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}
