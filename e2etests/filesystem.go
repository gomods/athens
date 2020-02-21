// +build e2etests

package e2etests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func setupTestRepo(repoPath, repoURL string) {
	os.RemoveAll(repoPath)
	cmd := exec.Command("git",
		"clone",
		repoURL,
		repoPath)

	cmd.Run()
}

func chmodR(path string, mode os.FileMode) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			os.Chmod(name, mode)
		}
		return err
	})
}

func cleanGoCache(env []string) error {
	cmd := exec.Command("go", "clean", "--modcache")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to clear go cache: %v - %s", err, string(output))
	}
	return nil
}
