package goget

import (
	"context"
	"os"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gomods/athens/pkg/config/env"
	cerrors "github.com/gomods/athens/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name    string
	mod     string
	version string
}

// TODO(marwan): we should create Test Repos under github.com/gomods
// so we can get reproducible results from live VCS repos.
// For now, I cannot test that github.com/pkg/errors returns v0.8.0
// from goget.Latest, because they could very well introduce a new tag
// in the near future.
var tt = []testCase{
	{"basic list", "github.com/pkg/errors", "latest"},
	{"list non tagged", "github.com/marwan-at-work/gowatch", "latest"},
	{"list vanity", "golang.org/x/tools", "latest"},
}

func TestList(t *testing.T) {
	dp := New()
	ctx := context.Background()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dp.List(ctx, tc.mod) // TODO ensure list is correct per TODO above.
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVersionBinPathError(t *testing.T) {
	goPath := env.GoBinPath()
	os.Setenv("GO_BINARY_PATH", "some_invalid_path")
	defer os.Setenv("GO_BINARY_PATH", goPath)
	envy.Reload()
	defer envy.Reload()
	const vop cerrors.Op = "module.getSources"

	dp := New()

	v, err := dp.Version(context.Background(), "mod", "version")

	assert.Nil(t, v)
	cerr, ok := err.(cerrors.Error)
	assert.True(t, ok)
	assert.EqualError(t, cerr.Err, "Invalid go binary: exit status 1")
}
