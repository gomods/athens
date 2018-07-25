package module

import "fmt"

// ErrModuleExcluded is error returned when processing of error is skipped
// due to filtering rules
type ErrModuleExcluded struct {
	module string
}

func (e *ErrModuleExcluded) Error() string {
	return fmt.Sprintf("Module %s is excluded", e.module)
}

// NewErrModuleExcluded creates new ErrModuleExcluded
func NewErrModuleExcluded(module string) error {
	return &ErrModuleExcluded{module: module}
}

// ErrModuleAlreadyFetched is an error returned when you try to fetch the same module@version
// more than once
type ErrModuleAlreadyFetched struct {
	module  string
	version string
}

// Error is the error interface implementation
func (e *ErrModuleAlreadyFetched) Error() string {
	return fmt.Sprintf("%s@%s was already fetched", e.module, e.version)
}

// NewErrModuleAlreadyFetched returns a new ErrModuleAlreadyFetched
func NewErrModuleAlreadyFetched(mod, ver string) error {
	return &ErrModuleAlreadyFetched{module: mod, version: ver}
}
