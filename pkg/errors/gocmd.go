package errors

import (
	"strings"
)

// IsRepoNotFoundErr checks if a repo has been found or not and reports an error accordingly
func IsRepoNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), "remote: Repository not found")
}
