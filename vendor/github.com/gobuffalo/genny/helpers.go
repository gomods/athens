package genny

import (
	"path/filepath"
	"strings"

	"github.com/gobuffalo/envy"
)

func exts(f File) []string {
	var exts []string

	name := f.Name()
	ext := filepath.Ext(name)

	for ext != "" {
		exts = append([]string{ext}, exts...)
		name = strings.TrimSuffix(name, ext)
		ext = filepath.Ext(name)
	}
	return exts
}

// HasExt checks if a file has a particular extension
func HasExt(f File, ext string) bool {
	if ext == "*" {
		return true
	}
	for _, x := range exts(f) {
		if x == ext {
			return true
		}
	}
	return false
}

// StripExt from a File and return a new one
func StripExt(f File, ext string) File {
	name := f.Name()
	name = strings.Replace(name, ext, "", -1)
	return NewFile(name, f)
}

func GoBin() string {
	return envy.Get("GO_BIN", "go")
}
