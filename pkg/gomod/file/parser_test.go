package file

import (
	"os"
	"strings"
	"testing"
)

func TestFileParser_ModuleName(t *testing.T) {
	expectedModuleName := "arschles.com/testmodule"

	parser := NewFileParser("../../../testmodule/go.mod")

	got, err := parser.ModuleName()
	if err != nil {
		t.Fatalf("Expected to find a module name... Got %v", err)
	}

	if !strings.EqualFold(got, expectedModuleName) {
		t.Fatalf(`Module names do not match.. \n 
Expected %s .. Got %s`, expectedModuleName, got)
	}
}

func TestFileParser_FileNotFound(t *testing.T) {
	parser := NewFileParser("../../../go.mod")

	_, err := parser.ModuleName()
	if !os.IsNotExist(err) {
		t.Fatalf("file does not exist but got %v instead", err)
	}
}
