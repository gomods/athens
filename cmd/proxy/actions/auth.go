package actions

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// initializeAuthFile checks if provided auth file is at a pre-configured path
// and moves to home directory -- note that this will override whatever
// .netrc/.hgrc file you have in your home directory.
func initializeAuthFile(path string) {
	if path == "" {
		return
	}

	fileBts, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("could not read %s: %v", path, err)
	}

	hdir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("could not get homedir: %v", err)
	}

	fileName := transformAuthFileName(filepath.Base(path))
	rcp := filepath.Join(hdir, fileName)
	if err := ioutil.WriteFile(rcp, fileBts, 0600); err != nil {
		log.Fatalf("could not write to file: %v", err)
	}
}

// netrcFromToken takes a github token and creates a .netrc
// file for you, overriding whatever might be already there.
func netrcFromToken(tok string) {
	fileContent := fmt.Sprintf("machine github.com login %s\n", tok)
	hdir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("netrcFromToken: could not get homedir: %v", err)
	}
	rcp := filepath.Join(hdir, getNetrcFileName())
	if err := ioutil.WriteFile(rcp, []byte(fileContent), 0600); err != nil {
		log.Fatalf("netrcFromToken: could not write to file: %v", err)
	}
}

func transformAuthFileName(authFileName string) string {
	if root := strings.TrimLeft(authFileName, "._"); root == "netrc" {
		return getNetrcFileName()
	}
	return authFileName
}

func getNetrcFileName() string {
	if runtime.GOOS == "windows" {
		return "_netrc"
	}
	return ".netrc"
}
