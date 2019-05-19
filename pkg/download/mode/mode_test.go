package mode

import (
	"testing"
)

var testCases = []struct {
	name         string
	file         *DownloadFile
	input        string
	expectedMode Mode
	expectedURL  string
}{
	{
		name:         "sync",
		file:         &DownloadFile{Mode: Sync},
		input:        "github.com/gomods/athens",
		expectedMode: Sync,
	},
	{
		name:         "redirect",
		file:         &DownloadFile{Mode: Redirect, DownloadURL: "gomods.io"},
		input:        "github.com/gomods/athens",
		expectedMode: Redirect,
		expectedURL:  "gomods.io",
	},
	{
		name: "pattern match",
		file: &DownloadFile{
			Mode: Sync,
			Paths: []*DownloadPath{
				{Pattern: "github.com/gomods/*", Mode: None},
			},
		},
		input:        "github.com/gomods/athens",
		expectedMode: None,
	},
	{
		name: "pattern fallback",
		file: &DownloadFile{
			Mode: Sync,
			Paths: []*DownloadPath{
				{Pattern: "github.com/gomods/*", Mode: None},
			},
		},
		input:        "github.com/athens-artifacts/maturelib",
		expectedMode: Sync,
	},
	{
		name: "pattern redirect",
		file: &DownloadFile{
			Mode: Sync,
			Paths: []*DownloadPath{
				{
					Pattern:     "github.com/gomods/*",
					Mode:        AsyncRedirect,
					DownloadURL: "gomods.io"},
			},
		},
		input:        "github.com/gomods/athens",
		expectedMode: AsyncRedirect,
		expectedURL:  "gomods.io",
	},
	{
		name: "redirect fallback",
		file: &DownloadFile{
			Mode:        Redirect,
			DownloadURL: "proxy.golang.org",
			Paths: []*DownloadPath{
				{
					Pattern:     "github.com/gomods/*",
					Mode:        AsyncRedirect,
					DownloadURL: "gomods.io",
				},
			},
		},
		input:        "github.com/athens-artifacts/maturelib",
		expectedMode: Redirect,
		expectedURL:  "proxy.golang.org",
	},
}

func TestMode(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			givenMode := tc.file.Match(tc.input)
			if givenMode != tc.expectedMode {
				t.Fatalf("expected matched mode to be %q but got %q", tc.expectedMode, givenMode)
			}
			givenURL := tc.file.URL(tc.input)
			if givenURL != tc.expectedURL {
				t.Fatalf("expected matched DownloadURL to be %q but got %q", tc.expectedURL, givenURL)
			}
		})
	}
}
