package github

import (
	"testing"
)

func Test_Download(t *testing.T) {
	owner := "bketelsen"
	repo := "captainhook"
	version := "v0.1.8"

	fetcher, err := NewGitFetcher(owner, repo, version)
	if err != nil {
		t.Error(err)
	}

	path, err := fetcher.Fetch()
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Error("path null")
	}
	t.Log(path)
	err = fetcher.Clear()
	if err != nil {
		t.Error(err)
	}
}
