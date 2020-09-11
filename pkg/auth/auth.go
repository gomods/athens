package auth

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/gomods/athens/pkg/errors"
)

type authkey struct{}

// BasicAuth is the embedded credentials in a context
type BasicAuth struct {
	User, Password string
}

// SetAuthInContext sets the auth value in context
func SetAuthInContext(ctx context.Context, auth BasicAuth) context.Context {
	return context.WithValue(ctx, authkey{}, auth)
}

// FromContext retrieves the auth value
func FromContext(ctx context.Context) (BasicAuth, bool) {
	auth, ok := ctx.Value(authkey{}).(BasicAuth)
	return auth, ok
}

// WriteNETRC writes the netrc file to the specified directory
func WriteNETRC(path, host, user, password string) error {
	const op errors.Op = "auth.WriteNETRC"
	fileContent := fmt.Sprintf("machine %s login %s password %s\n", host, user, password)
	if err := ioutil.WriteFile(path, []byte(fileContent), 0600); err != nil {
		return errors.E(op, fmt.Errorf("netrcFromToken: could not write to file: %v", err))
	}
	return nil
}

// WriteTemporaryNETRC writes a netrc file to a temporary directory, returning
// the directory it was written to.
func WriteTemporaryNETRC(host, user, password string) (string, error) {
	const op errors.Op = "auth.WriteTemporaryNETRC"
	dir, err := ioutil.TempDir("", "netrcp")
	if err != nil {
		return "", errors.E(op, err)
	}
	rcp := filepath.Join(dir, GetNETRCFilename())
	err = WriteNETRC(rcp, host, user, password)
	if err != nil {
		return "", errors.E(op, err)
	}
	return dir, nil
}

// GetNETRCFilename returns the name of the netrc file
// according to the contextual platform
func GetNETRCFilename() string {
	if runtime.GOOS == "windows" {
		return "_netrc"
	}
	return ".netrc"
}
