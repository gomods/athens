package errors

import (
	"strings"
)

func IsRepoNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), "remote: Repository not found")
}
