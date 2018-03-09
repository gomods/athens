package zip

import (
	"archive/zip"
	"fmt"

	"github.com/gomods/athens/pkg/parser"
)

const (
	gomodFilename = "go.mod"
)

func NewZipParser(reader zip.ReadCloser) parser.GomodParser {
	return zipParser{zipReader: reader}
}

type zipParser struct {
	zipReader zip.ReadCloser
}

func (p zipParser) ModuleName() (string, error) {
	defer p.zipReader.Close()

	var file *zip.File
	for _, f := range p.zipReader.File {
		if f.Name != gomodFilename {
			continue
		}

		file = f
		break
	}

	if file == nil {
		return "", fmt.Errorf("go.mod not found")
	}

	fileReader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer fileReader.Close()

	return parser.Parse(fileReader)
}
