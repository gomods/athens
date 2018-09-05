package download

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"

	athenser "github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
)

type listMergeTest struct {
	name        string
	module      string
	goVersions  []string
	goErr       error
	strVersions []string
	strErr      error
	expected    []string
	expectedErr error
}

const testOp athenser.Op = "protocol.List"

var listMergeTests = []listMergeTest{
	{
		name:        "go list full and storage full",
		module:      "happy tags",
		goVersions:  []string{"v1.0.0", "v1.0.2", "v1.0.3"},
		goErr:       nil,
		strVersions: []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		expected:    []string{"v1.0.0", "v1.0.1", "v1.0.2", "v1.0.3"},
		expectedErr: nil,
	},
	{
		name:        "go list full and storage empty",
		module:      "happy tags",
		goVersions:  []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		goErr:       nil,
		strVersions: []string{},
		strErr:      athenser.E(testOp, athenser.M("happy tags"), athenser.KindNotFound),
		expected:    []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		expectedErr: nil,
	},
	{
		name:        "go list repo not found and storage full",
		module:      "happy tags",
		goVersions:  nil,
		goErr:       errors.New("remote: Repository not found"),
		strVersions: []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		strErr:      nil,
		expected:    []string{"v1.0.0", "v1.0.1", "v1.0.2"},
		expectedErr: nil,
	},
	{
		name:        "go list repo not found and storage empty",
		module:      "happy tags",
		goVersions:  nil,
		goErr:       errors.New("remote: Repository not found"),
		strVersions: nil,
		strErr:      athenser.E(testOp, athenser.M("happy tags"), athenser.KindNotFound),
		expected:    nil,
		expectedErr: athenser.E(testOp, athenser.M("happy tags"), athenser.KindNotFound),
	},
	{
		name:        "unexpected go err",
		module:      "happy tags",
		goVersions:  nil,
		goErr:       errors.New("unexpected error"),
		strVersions: []string{"1.1.1"},
		strErr:      nil,
		expected:    nil,
		expectedErr: athenser.E(testOp, errors.New("unexpected error")),
	},
	{
		name:        "unexpected storage err",
		module:      "happy tags",
		goVersions:  []string{"1.1.1"},
		goErr:       nil,
		strVersions: nil,
		strErr:      errors.New("unexpected error"),
		expected:    nil,
		expectedErr: athenser.E(testOp, errors.New("unexpected error")),
	},
}

func TestListMerge(t *testing.T) {
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

	newLister := func(versions []string, err error) Lister {
		return func(mod string) (*storage.RevInfo, []string, error) {
			return nil, versions, err
		}
	}
	for _, tc := range listMergeTests {
		t.Run(tc.name, func(t *testing.T) {
			bts := []byte("123")
			for _, v := range tc.strVersions {
				s.Save(ctx, tc.module, v, bts, ioutil.NopCloser(bytes.NewReader(bts)), bts)
			}
			defer clearStorage(s, tc.module, tc.strVersions)
			dp := New(&Opts{s, nil, newLister(tc.goVersions, tc.goErr)})
			list, _ := dp.List(ctx, tc.module)

			if ok := testEq(tc.expected, list); !ok {
				t.Fatalf("expected list: %v, got: %v", tc.expected, list)
			}
		})
	}
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
