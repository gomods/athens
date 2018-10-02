package actions

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// initializeDotFile checks if provided dot file is at a pre-configured path
// and moves to home directory -- note that this will override whatever
// dot file you have in your home directory.
func initializeDotFile(path string) {
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
	rcp := filepath.Join(hdir, filepath.Base(path))
	ioutil.WriteFile(rcp, fileBts, 0666)
}
