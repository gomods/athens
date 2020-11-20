package module

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// prepareEnv will return all the appropriate
// environment variables for a Go Command to run
// successfully (such as GOPATH, GOCACHE, PATH etc)
func prepareEnv(gopath, homedir string, envVars []string) []string {
	gopathEnv := fmt.Sprintf("GOPATH=%s", gopath)
	cacheEnv := fmt.Sprintf("GOCACHE=%s", filepath.Join(gopath, "cache"))
	disableCgo := "CGO_ENABLED=0"
	enableGoModules := "GO111MODULE=on"
	cmdEnv := []string{
		gopathEnv,
		cacheEnv,
		disableCgo,
		enableGoModules,
	}
	cmdEnv = append(cmdEnv, withHomeDir(homedir)...)
	keys := []string{
		"PATH",
		"GIT_SSH",
		"GIT_SSH_COMMAND",
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"NO_PROXY",
		// Need to also check the lower case version of just these three env variables.
		"http_proxy",
		"https_proxy",
		"no_proxy",
	}
	if runtime.GOOS == "windows" {
		windowsSpecificKeys := []string{
			"SystemRoot",
			"ALLUSERSPROFILE",
			"HOMEDRIVE",
			"HOMEPATH",
		}
		keys = append(keys, windowsSpecificKeys...)
	}
	for _, key := range keys {
		// Prepend only if environment variable is present.
		if v, ok := os.LookupEnv(key); ok {
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", key, v))
		}
	}
	cmdEnv = append(cmdEnv, envVars...)

	if sshAuthSockVal, hasSSHAuthSock := os.LookupEnv("SSH_AUTH_SOCK"); hasSSHAuthSock {
		// Verify that the ssh agent unix socket exists and is a unix socket.
		st, err := os.Stat(sshAuthSockVal)
		if err == nil && st.Mode()&os.ModeSocket != 0 {
			sshAuthSock := fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSockVal)
			cmdEnv = append(cmdEnv, sshAuthSock)
		}
	}
	return cmdEnv
}

func withHomeDir(dir string) []string {
	key := "HOME"
	if runtime.GOOS == "windows" {
		key = "USERPROFILE"
	}
	if dir != "" {
		return []string{key + "=" + dir}
	}
	val, ok := os.LookupEnv(key)
	if ok {
		return []string{key + "=" + val}
	}
	return []string{}
}
