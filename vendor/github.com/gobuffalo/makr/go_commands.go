package makr

import (
	"os/exec"

	"github.com/gobuffalo/envy"
)

func turnOffMods(fn func()) {
	gm := envy.Get("GO111MODULE", "off")
	defer func() {
		if err := envy.MustSet("GO111MODULE", gm); err != nil {
			panic(err)
		}
	}()
	if err := envy.MustSet("GO111MODULE", "off"); err != nil {
		panic(err)
	}
	fn()
}

// GoInstall compiles and installs packages and dependencies
func GoInstall(pkg string, opts ...string) *exec.Cmd {
	var cmd *exec.Cmd
	turnOffMods(func() {
		args := append([]string{"install"}, opts...)
		args = append(args, pkg)
		cmd = exec.Command(envy.Get("GO_BIN", "go"), args...)
	})
	return cmd
}

// GoGet downloads and installs packages and dependencies
func GoGet(pkg string, opts ...string) *exec.Cmd {
	var cmd *exec.Cmd
	turnOffMods(func() {
		args := append([]string{"get"}, opts...)
		args = append(args, pkg)
		cmd = exec.Command(envy.Get("GO_BIN", "go"), args...)
	})
	return cmd
}

// GoFmt is command that will use `goimports` if available,
// or fail back to `gofmt` otherwise.
func GoFmt(files ...string) *exec.Cmd {
	if len(files) == 0 {
		files = []string{"."}
	}
	c := "gofmt"
	_, err := exec.LookPath("goimports")
	if err == nil {
		c = "goimports"
	}
	_, err = exec.LookPath("gofmt")
	if err != nil {
		return exec.Command("echo", "could not find gofmt or goimports")
	}
	args := []string{"-w"}
	args = append(args, files...)
	return exec.Command(c, args...)
}
