package memory

import "github.com/gomods/athens/pkg/eventlog"

// Registry is a pointer registry for olympus server event logs
type Registry struct {
	registry map[string]string
}

// NewRegistry creates a pointer registry
func NewRegistry() *Registry {
	memRegistry := make(map[string]string)
	return &Registry{memRegistry}
}

// LookupPointer returns the event log pointer for a given deployment ID
func (r *Registry) LookupPointer(deploymentID string) (string, error) {
	var p string
	if p, ok := r.registry[deploymentID]; !ok {
		return p, eventlog.ErrDeploymentNotFound
	}
	return p, nil
}

// TODO: SetPointer
