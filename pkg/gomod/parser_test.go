package parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_ScannerError(t *testing.T) {

	_, err := Parse(mockScanner{})
	if err == nil {
		t.Fatal("expected an error to have occurred")
	}
}

type mockScanner struct {
}

func (m mockScanner) Read(p []byte) (int, error) {
	return 0, errors.New("gomod : an error occurred while ")
}

func TestParse_checkVersion(t *testing.T) {
	cases := []struct {
		line     string
		expected bool
	}{
		{"", false},
		{"github.com/golang/go", false},
		{"module ", false},
		{"module github.com/gomods/athens", false},
		{`module "github.com/gomods/athens"`, true},
	}

	for _, val := range cases {
		got, truthy := checkVersion(val.line, re)

		if val.expected {
			if strings.EqualFold(got, val.line) {
				t.Fatalf("Module names do not match... Expected %s, Got %s",
					val.line, got)
			}

			continue
		}

		if truthy != val.expected {
			t.Fatalf(`Expected "%s" to have %v ... Got %v instead`,
				val.line, val.expected, truthy)
		}
	}
}
