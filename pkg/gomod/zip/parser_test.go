package zip

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipParser_ModuleName(t *testing.T) {

	moduleName := "arschles.com/testmodule"

	fileName, err := zipTestModule(t)
	if err != nil {
		t.Fatalf("an error occurred while zipping the test module.. %v", err)
	}

	rc, err := zip.OpenReader(fileName)
	if err != nil {
		t.Fatalf("an error occured while opening zip file... %v", err)
	}

	parser := NewZipParser(*rc)

	got, err := parser.ModuleName()
	if err != nil {
		t.Fatalf("Expected to find a module name... Got %v", err)
	}

	if !strings.EqualFold(got, moduleName) {
		t.Fatalf(`Module names do not match.. \n 
Expected %s .. Got %s`, moduleName, got)
	}
}

func zipTestModule(t *testing.T) (target string, err error) {

	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("an error occurred while creating temporary file")
	}

	target = file.Name()

	archive := zip.NewWriter(file)
	defer func() {
		archive.Close()
		file.Close()
	}()

	src := "../../../testmodule"

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return
}
