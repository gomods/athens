package stash

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/storage"
)

type stashTest struct {
	name             string
	ver              string // the given version
	modVer           string // the version module.Fetcher returns
	shouldCallExists bool   // whether storage should be checked before saving
	existsResponse   bool   // the response of storage.Exists if it's called
	shouldCallSave   bool   // whether save or not should be called
}

var stashTests = [...]stashTest{
	{
		name:             "non semver",
		ver:              "master",
		modVer:           "v1.2.3",
		shouldCallExists: true,
		existsResponse:   false,
		shouldCallSave:   true,
	},
	{
		name:             "no storage override",
		ver:              "master",
		modVer:           "v1.2.3",
		shouldCallExists: true,
		existsResponse:   true,
		shouldCallSave:   false,
	},
	{
		name:             "equal semver",
		ver:              "v2.0.0",
		modVer:           "v2.0.0",
		shouldCallExists: false,
		existsResponse:   false,
		shouldCallSave:   true,
	},
}

func TestStash(t *testing.T) {
	for _, testCase := range stashTests {
		t.Run(testCase.name, func(t *testing.T) {
			var ms mockStorage
			ms.existsResponse = testCase.existsResponse
			var mf mockFetcher
			mf.ver = testCase.modVer

			s := New(&mf, &ms)
			newVersion, err := s.Stash(context.Background(), "module", testCase.ver)
			if err != nil {
				t.Fatal(err)
			}
			if newVersion != testCase.modVer {
				t.Fatalf("expected Stash to return %v from module.Fetcher but got %v", testCase.modVer, newVersion)
			}
			if testCase.shouldCallExists != ms.existsCalled {
				t.Fatalf("expected a call to storage.Exists to be %v but got %v", testCase.shouldCallExists, ms.existsCalled)
			}
			if testCase.shouldCallSave {
				if ms.givenVersion != testCase.modVer {
					t.Fatalf("expected storage.Save to be called with version %v but got %v", testCase.modVer, ms.givenVersion)
				}
			} else if ms.saveCalled {
				t.Fatalf("expected save not to be called")
			}
		})
	}
}

type mockStorage struct {
	storage.Backend
	existsCalled   bool
	saveCalled     bool
	givenVersion   string
	existsResponse bool
}

func (ms *mockStorage) Save(ctx context.Context, module, version string, mod []byte, zip storage.Zip, info []byte) error {
	ms.saveCalled = true
	ms.givenVersion = version
	return nil
}

func (ms *mockStorage) Exists(ctx context.Context, mod, ver string) (bool, error) {
	ms.existsCalled = true
	return ms.existsResponse, nil
}

type mockFetcher struct {
	ver string
}

func (mf *mockFetcher) Fetch(ctx context.Context, mod, ver string) (*storage.Version, error) {
	return &storage.Version{
		Info:   []byte("info"),
		Mod:    []byte("gomod"),
		Zip:    storage.Zip{ioutil.NopCloser(strings.NewReader("zipfile")), 7},
		Semver: mf.ver,
	}, nil
}
