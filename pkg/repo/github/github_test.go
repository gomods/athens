package github

import (
	"os"
	"testing"
)

func Test_Download(t *testing.T) {
	o := "bketelsen"
	r := "captainhook"
	v := "v0.1.8"

	fetcher, err := NewGitCrawler(o, r, v)
	if err != nil {
		t.Error(err)
	}

	path, err := fetcher.DownloadRepo()
	if err != nil {
		t.Error(err)
	}
	if path == "" {
		t.Error("path null")
	}
	t.Log(path)
	os.RemoveAll(path)
}
