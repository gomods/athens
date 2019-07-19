package module

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/afero"
)

type goTool struct {
	goBin   string
	goProxy string
	goPath  string
}

type goRuntime struct {
	fs      afero.Fs
	root    string
	workDir string

	stdout io.ReadWriter
	stderr io.ReadWriter
	env    []string

	goTool
}

func prepareRuntime(fs afero.Fs, goTool goTool, workDir string) (*goRuntime, error) {
	goPath, err := afero.TempDir(fs, "", "athens")
	if err != nil {
		return nil, err
	}
	if workDir == "" {
		workDir = goPath
	} else {
		workDir = filepath.Join(goPath, workDir)
		if err := fs.MkdirAll(workDir, os.ModeDir|os.ModePerm); err != nil {
			ClearFiles(fs, goPath)
			return nil, err
		}
	}
	goTool.goPath = goPath
	env := PrepareEnv(goTool)
	return &goRuntime{
		fs:      fs,
		root:    goPath,
		workDir: workDir,

		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
		env:    env,

		goTool: goTool,
	}, nil
}

func (gr *goRuntime) run(args ...string) error {
	cmd := exec.Command(gr.goBin, args...)
	cmd.Dir = gr.workDir
	cmd.Stdout = gr.stdout
	cmd.Stderr = gr.stderr
	cmd.Env = gr.env
	return cmd.Run()
}

func (gr *goRuntime) clean() error {
	return ClearFiles(gr.fs, gr.root)
}

// PrepareEnv will return all the appropriate
// environment variables for a Go Command to run
// successfully (such as GOPATH, GOCACHE, PATH etc)
func PrepareEnv(goTool goTool) []string {
	pathEnv := fmt.Sprintf("PATH=%s", os.Getenv("PATH"))
	homeEnv := fmt.Sprintf("HOME=%s", os.Getenv("HOME"))
	httpProxy := fmt.Sprintf("HTTP_PROXY=%s", os.Getenv("HTTP_PROXY"))
	httpsProxy := fmt.Sprintf("HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY"))
	noProxy := fmt.Sprintf("NO_PROXY=%s", os.Getenv("NO_PROXY"))
	// need to also check the lower case version of just these three env variables
	httpProxyLower := fmt.Sprintf("http_proxy=%s", os.Getenv("http_proxy"))
	httpsProxyLower := fmt.Sprintf("https_proxy=%s", os.Getenv("https_proxy"))
	noProxyLower := fmt.Sprintf("no_proxy=%s", os.Getenv("no_proxy"))
	gopathEnv := fmt.Sprintf("GOPATH=%s", goTool.goPath)
	goProxyEnv := fmt.Sprintf("GOPROXY=%s", goTool.goProxy)
	cacheEnv := fmt.Sprintf("GOCACHE=%s", filepath.Join(goTool.goPath, "cache"))
	gitSSH := fmt.Sprintf("GIT_SSH=%s", os.Getenv("GIT_SSH"))
	gitSSHCmd := fmt.Sprintf("GIT_SSH_COMMAND=%s", os.Getenv("GIT_SSH_COMMAND"))
	disableCgo := "CGO_ENABLED=0"
	enableGoModules := "GO111MODULE=on"
	cmdEnv := []string{
		pathEnv,
		homeEnv,
		gopathEnv,
		goProxyEnv,
		cacheEnv,
		disableCgo,
		enableGoModules,
		httpProxy,
		httpsProxy,
		noProxy,
		httpProxyLower,
		httpsProxyLower,
		noProxyLower,
		gitSSH,
		gitSSHCmd,
	}

	if sshAuthSockVal, hasSSHAuthSock := os.LookupEnv("SSH_AUTH_SOCK"); hasSSHAuthSock {
		// Verify that the ssh agent unix socket exists and is a unix socket.
		st, err := os.Stat(sshAuthSockVal)
		if err == nil && st.Mode()&os.ModeSocket != 0 {
			sshAuthSock := fmt.Sprintf("SSH_AUTH_SOCK=%s", sshAuthSockVal)
			cmdEnv = append(cmdEnv, sshAuthSock)
		}
	}

	// add Windows specific ENV VARS
	if runtime.GOOS == "windows" {
		cmdEnv = append(cmdEnv, fmt.Sprintf("USERPROFILE=%s", os.Getenv("USERPROFILE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("SystemRoot=%s", os.Getenv("SystemRoot")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("ALLUSERSPROFILE=%s", os.Getenv("ALLUSERSPROFILE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("HOMEDRIVE=%s", os.Getenv("HOMEDRIVE")))
		cmdEnv = append(cmdEnv, fmt.Sprintf("HOMEPATH=%s", os.Getenv("HOMEPATH")))
	}

	return cmdEnv
}

type runtimeClean func() error
