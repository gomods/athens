package errors

import "strings"

// IsNoChildProcessesErr returns true for an error from
// exec.Command().Run() that can be safely ignored.
// Reference: https://github.com/slimtoolkit/slim/blob/79b63a80c10083ece51be0ef1fd1e7c090ff6346/pkg/util/errutil/errutil.go#L95-L110
func IsNoChildProcessesErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "wait: no child processes") ||
		strings.Contains(err.Error(), "waitid: no child processes")
}
