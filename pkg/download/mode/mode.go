package mode

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
)

// Mode specifies the behavior of what to do
// when a module is not found in storage.
type Mode string

// DownloadMode constants. For more information see config.dev.toml
const (
	Sync            Mode = "sync"
	Async           Mode = "async"
	Redirect        Mode = "redirect"
	AsyncRedirect   Mode = "async_redirect"
	None            Mode = "none"
	downloadModeErr      = "download mode is not set"
	invalidModeErr       = "unrecognized download mode: %s"
)

// DownloadFile represents a custom HCL format of
// how to handle module@version requests that are
// not found in storage.
type DownloadFile struct {
	Mode        Mode            `hcl:"mode"`
	DownloadURL string          `hcl:"downloadURL"`
	Paths       []*DownloadPath `hcl:"download,block"`
}

// DownloadPath represents a custom Mode for
// a matching path.
type DownloadPath struct {
	Pattern     string `hcl:"pattern,label"`
	Mode        Mode   `hcl:"mode"`
	DownloadURL string `hcl:"downloadURL,optional"`
}

// NewFile takes a mode and returns a DownloadFile.
// Mode can be one of the constants declared above
// or a custom HCL file. To pass a custom HCL file,
// you can either point to a file path by passing
// file:/path/to/file OR custom:<base64-encoded-hcl>
// directly.
func NewFile(m Mode, downloadURL string) (*DownloadFile, error) {
	const op errors.Op = "downloadMode.NewFile"

	if m == "" {
		return nil, errors.E(op, downloadModeErr)
	}

	if strings.HasPrefix(string(m), "file:") {
		filePath := string(m[5:])
		bts, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return parseFile(bts)
	} else if strings.HasPrefix(string(m), "custom:") {
		bts, err := base64.StdEncoding.DecodeString(string(m[7:]))
		if err != nil {
			return nil, err
		}
		return parseFile(bts)
	}

	switch m {
	case Sync, Async, Redirect, AsyncRedirect, None:
		return &DownloadFile{Mode: m, DownloadURL: downloadURL}, nil
	default:
		return nil, errors.E(op, errors.KindBadRequest, fmt.Sprintf(invalidModeErr, m))
	}
}

// parseFile parses an HCL file according to the
// DownloadMode spec.
func parseFile(file []byte) (*DownloadFile, error) {
	const op errors.Op = "downloadmode.parseFile"
	f, dig := hclparse.NewParser().ParseHCL(file, "config.hcl")
	if dig.HasErrors() {
		return nil, errors.E(op, dig.Error())
	}
	var df DownloadFile
	dig = gohcl.DecodeBody(f.Body, nil, &df)
	if dig.HasErrors() {
		return nil, errors.E(op, dig.Error())
	}
	if err := df.validate(); err != nil {
		return nil, errors.E(op, err)
	}
	return &df, nil
}

func (d *DownloadFile) validate() error {
	const op errors.Op = "downloadMode.validate"
	for _, p := range d.Paths {
		switch p.Mode {
		case Sync, Async, Redirect, AsyncRedirect, None:
		default:
			return errors.E(op, fmt.Errorf("unrecognized mode for %v: %v", p.Pattern, p.Mode))
		}
	}
	return nil
}

// Match returns the Mode that matches the given
// module. A pattern is prioritized by order in
// which it appears in the HCL file, while the
// default Mode will be returned if no patterns
// exist or match.
func (d *DownloadFile) Match(mod string) Mode {
	for _, p := range d.Paths {
		if paths.MatchesPattern(p.Pattern, mod) {
			return p.Mode
		}
	}
	return d.Mode
}

// URL returns the redirect URL that applies
// to the given module. If no pattern matches,
// the top level downloadURL is returned.
func (d *DownloadFile) URL(mod string) string {
	for _, p := range d.Paths {
		if paths.MatchesPattern(p.Pattern, mod) {
			if p.DownloadURL != "" {
				return p.DownloadURL
			}
		}
	}
	return d.DownloadURL
}
