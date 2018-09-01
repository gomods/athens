package download

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/gomods/athens/pkg/storage/mem"

	"github.com/gomods/athens/pkg/storage"
	"golang.org/x/sync/errgroup"
)

type testMod struct {
	mod, ver string
}

var mods = []testMod{
	{"github.com/athens-artifacts/no-tags", "v0.0.2"},
	{"github.com/athens-artifacts/happy-path", "v0.0.0-20180803035119-e4e0177efdb5"},
	{"github.com/athens-artifacts/samplelib", "v1.0.0"},
}

func TestDownloadProtocol(t *testing.T) {
	s, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	dp := New(&mockProtocol{}, s, 2)
	ctx := context.Background()

	var eg errgroup.Group
	for i := 0; i < len(mods); i++ {
		m := mods[i]
		eg.Go(func() error {
			_, err := dp.GoMod(ctx, m.mod, m.ver)
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range mods {
		bts, err := dp.GoMod(ctx, m.mod, m.ver)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(bts, []byte(m.mod+"@"+m.ver)) {
			t.Fatalf("unexpected gomod content: %s", bts)
		}
	}
}

type listTest struct {
	name          string
	module        string
	gogetVersions []string
	strVersions   []string
	expected      []string
}

var listTests = []listTest{
	{
		name:          "go list full and storage full",
		module:        "happy tags",
		gogetVersions: []string{"v1.0.0", "v1.0.2", "v1.0.3"},
		strVersions:   []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		expected:      []string{"v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3"},
	},
	{
		name:          "go list full and storage empty",
		module:        "happy tags",
		gogetVersions: []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		strVersions:   []string{},
		expected:      []string{"v1.0.0", "v1.0.1", "v1.0.2"},
	},
	{
		name:          "go list empty and storage full",
		module:        "happy tags",
		gogetVersions: nil,
		strVersions:   []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		expected:      []string{"v1.0.0", "v1.0.1", "v1.0.2"},
	},
	{
		name:          "go list empty and storage empty",
		module:        "happy tags",
		gogetVersions: nil,
		strVersions:   []string{},
		expected:      nil,
	},
}

func TestList(t *testing.T) {
	ctx := context.Background()
	s, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	clearStorage := func(st storage.Backend, module string, versions []string) {
		for _, v := range versions {
			s.Delete(ctx, module, v)
		}
	}

	for _, tc := range listTests {
		t.Run(tc.name, func(t *testing.T) {
			bts := []byte("123")
			for _, v := range tc.strVersions {
				s.Save(ctx, tc.module, v, bts, ioutil.NopCloser(bytes.NewReader(bts)), bts)
			}
			defer clearStorage(s, tc.module, tc.strVersions)
			dp := New(&mockProtocol{list: tc.gogetVersions}, s, 1)
			list, _ := dp.List(ctx, tc.module)

			if ok := testEq(tc.expected, list); !ok {
				t.Fatalf("expected list: %v, got: %v", tc.expected, list)
			}
		})
	}
}

type mockProtocol struct {
	Protocol
	list []string
}

// Info implements GET /{module}/@v/{version}.info
func (m *mockProtocol) Info(ctx context.Context, mod, ver string) ([]byte, error) {
	return []byte(mod + "@" + ver), nil
}

func (m *mockProtocol) Version(ctx context.Context, mod, ver string) (*storage.Version, error) {
	bts := []byte(mod + "@" + ver)
	return &storage.Version{
		Mod:  bts,
		Info: bts,
		Zip:  ioutil.NopCloser(bytes.NewReader(bts)),
	}, nil
}

func (m *mockProtocol) List(ctx context.Context, mod string) ([]string, error) {
	if m.list == nil {
		return nil, fmt.Errorf("list empty")
	}
	return m.list, nil
}

func testEq(a, b []string) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
