package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// initializeAuthFile checks if provided auth file is at a pre-configured path
// and moves to home directory -- note that this will override whatever
// .netrc/.hgrc file you have in your home directory.
func initializeAuthFile(path string) error {
	if path == "" {
		return nil
	}

	fileBts, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", path, err)
	}

	hdir, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("could not get homedir: %w", err)
	}

	fileName := transformAuthFileName(filepath.Base(path))
	rcp := filepath.Join(hdir, fileName)
	if err := os.WriteFile(rcp, fileBts, 0600); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

// netrcFromToken takes a github token and creates a .netrc
// file for you, overriding whatever might be already there.
func netrcFromToken(tok string) error {
	fileContent := fmt.Sprintf("machine github.com login %s\n", tok)
	hdir, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("could not get homedir: %w", err)
	}
	rcp := filepath.Join(hdir, getNETRCFilename())
	if err := os.WriteFile(rcp, []byte(fileContent), 0600); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}
	return nil
}

func transformAuthFileName(authFileName string) string {
	if root := strings.TrimLeft(authFileName, "._"); root == "netrc" {
		return getNETRCFilename()
	}
	return authFileName
}

func getNETRCFilename() string {
	if runtime.GOOS == "windows" {
		return "_netrc"
	}
	return ".netrc"
}
