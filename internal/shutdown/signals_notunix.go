//go:build !unix

package shutdown

import "os"

// GetSignals returns the appropriate signals to catch for a clean shutdown, dependent on the OS.
func GetSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
