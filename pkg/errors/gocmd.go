package errors

import (
	"strings"
)

// IsRepoNotFoundErr returns true if the Go command line
// hints at a repository not found.
func IsRepoNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), "remote: Repository not found")
}
