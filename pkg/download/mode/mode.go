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
	Sync          Mode = "sync"
	Async         Mode = "async"
	Redirect      Mode = "redirect"
	AsyncRedirect Mode = "async_redirect"
	None          Mode = "none"
)

// Validate ensures that m is a valid mode. If this function returns nil, you are
// guaranteed that m is valid
func (m Mode) Validate(op errors.Op) error {
	switch m {
	case Sync, Async, Redirect, AsyncRedirect, None:
		return nil
	default:
		return errors.Config(
			op,
			"mode",
			fmt.Sprintf("%s isn't a valid value.", m),
			"https://docs.gomods.io/configuration/download/",
		)
	}
}

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

func (d DownloadPath) validate(op errors.Op) error {
	if d.DownloadURL == "" && (d.Mode == Redirect || d.Mode == AsyncRedirect) {
		return errors.Config(
			op,
			fmt.Sprintf("DownloadURL (inside %s in the download file)", d.Pattern),
			"You must set a value when the download mode is 'redirect' or 'async_redirect'",
			"https://docs.gomods.io/configuration/download/",
		)
	}
	return nil
}

// NewFile takes a mode and returns a DownloadFile.
// Mode can be one of the constants declared above
// or a custom HCL file. To pass a custom HCL file,
// you can either point to a file path by passing
// file:/path/to/file OR custom:<base64-encoded-hcl>
// directly.
func NewFile(m Mode, downloadURL string) (*DownloadFile, error) {
	const downloadModeURL = "https://docs.gomods.io/configuration/download/"
	const op errors.Op = "downloadMode.NewFile"

	if err := m.Validate(op); err != nil {
		return nil, err
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

	retFile := &DownloadFile{Mode: m, DownloadURL: downloadURL}
	if err := retFile.validate(op); err != nil {
		return nil, err
	}
	return retFile, nil
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
	if err := df.validate(op); err != nil {
		return nil, errors.E(op, err)
	}
	return &df, nil
}

func (d *DownloadFile) validate(op errors.Op) error {
	for _, p := range d.Paths {
		if err := p.Mode.Validate(op); err != nil {
			return err
		}
		switch p.Mode {
		case Sync, Async, Redirect, AsyncRedirect, None:
		default:
			return errors.Config(
				op,
				fmt.Sprintf("mode (in pattern %v", p.Pattern),
				fmt.Sprintf("%s is unrecognized", p.Mode),
				"https://docs.gomods.io/configuration/download/",
			)
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
