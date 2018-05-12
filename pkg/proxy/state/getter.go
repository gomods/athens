package state

import "github.com/gomods/athens/pkg/proxy"

// Getter gets the state of the proxy
type Getter interface {
	Get() (proxy.State, error)
}
