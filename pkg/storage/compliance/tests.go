package compliance

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/stretchr/testify/require"
)

// RunTests takes a backend implementation and runs compliance tests
// against the interface.
func RunTests(t *testing.T, b storage.Backend, clearBackend func() error) {
	require.NoError(t, clearBackend(), "clearing backend failed")
	defer require.NoError(t, clearBackend(), "clearning backend failed")
	testNotFound(t, b)
	testList(t, b)
	testDelete(t, b)
	testGet(t, b)
	testCatalog(t, b)
}

// testNotFound ensures that a storage Backend
// returns a KindNotFound error when asking for
// non existing modules.
func testNotFound(t *testing.T, b storage.Backend) {
	mod, ver := "xxx", "yyy"
	ctx := context.Background()

	err := b.Delete(ctx, mod, ver)
	require.Error(t, err)
	require.Equal(t, errors.KindNotFound, errors.Kind(err))

	_, err = b.GoMod(ctx, mod, ver)
	require.Error(t, err)
	require.Equal(t, errors.KindNotFound, errors.Kind(err))

	_, err = b.Info(ctx, mod, ver)
	require.Error(t, err)
	require.Equal(t, errors.KindNotFound, errors.Kind(err))

	vs, err := b.List(ctx, mod)
	require.NoError(t, err)
	require.Equal(t, 0, len(vs))

	_, err = b.Zip(ctx, mod, ver)
	require.Error(t, err)
	require.Equal(t, errors.KindNotFound, errors.Kind(err))
}

// testList tests that a storage Backend returns
// the exact list of versions that are saved.
func testList(t *testing.T, b storage.Backend) {
	ctx := context.Background()

	modname := "listMod"
	versions := []string{"v1.1.0", "v1.2.0", "v1.3.0"}
	for _, version := range versions {
		mock := getMockModule()
		err := b.Save(
			ctx,
			modname,
			version,
			mock.Mod,
			mock.Zip,
			mock.Info,
		)
		require.NoError(t, err, "Save for storage failed")
	}
	defer func() {
		for _, ver := range versions {
			b.Delete(ctx, modname, ver)
		}
	}()
	retVersions, err := b.List(ctx, modname)
	require.NoError(t, err)
	require.Equal(t, versions, retVersions)
}

// testGet saves and retrieves a module successfully.
func testGet(t *testing.T, b storage.Backend) {
	ctx := context.Background()
	modname := "getTestModule"
	ver := "v1.2.3"
	mock := getMockModule()
	zipBts, _ := ioutil.ReadAll(mock.Zip)
	b.Save(ctx, modname, ver, mock.Mod, bytes.NewReader(zipBts), mock.Info)
	defer b.Delete(ctx, modname, ver)

	info, err := b.Info(ctx, modname, ver)
	require.NoError(t, err)
	require.Equal(t, mock.Info, info)

	mod, err := b.GoMod(ctx, modname, ver)
	require.NoError(t, err)
	require.Equal(t, mock.Mod, mod)

	zip, err := b.Zip(ctx, modname, ver)
	require.NoError(t, err)
	givenZipBts, err := ioutil.ReadAll(zip)
	require.NoError(t, err)
	require.Equal(t, zipBts, givenZipBts)
}

// testDelete tests that a module can be deleted from a
// storage Backend and the the Exists method returns false
// afterwards.
func testDelete(t *testing.T, b storage.Backend) {
	ctx := context.Background()
	modname := "deleteModule"
	version := fmt.Sprintf("%s%d", "delete", rand.Int())

	mock := getMockModule()
	err := b.Save(ctx, modname, version, mock.Mod, mock.Zip, mock.Info)
	require.NoError(t, err)

	err = b.Delete(ctx, modname, version)
	require.NoError(t, err)

	exists, err := b.Exists(ctx, modname, version)
	require.NoError(t, err)
	require.Equal(t, false, exists)
}

func testCatalog(t *testing.T, b storage.Backend) {
	ctx := context.Background()

	mock := getMockModule()
	zipBts, _ := ioutil.ReadAll(mock.Zip)
	modname := "testCatalogModule"

	for i := 0; i < 1005; i++ {
		ver := fmt.Sprintf("v1.2.%03d", i)
		b.Save(ctx, modname, ver, mock.Mod, bytes.NewReader(zipBts), mock.Info)
	}
	defer func() {
		for i := 0; i < 1005; i++ {
			ver := fmt.Sprintf("v1.2.03%d", i)
			b.Delete(ctx, modname, ver)
		}
	}()

	res, next, err := b.Catalog(ctx, "", 5)
	require.NoError(t, err)
	require.Equal(t, len(res), 5)

	res, next, err = b.Catalog(ctx, next, 50)
	require.NoError(t, err)
	require.Equal(t, len(res), 50)

	res, next, err = b.Catalog(ctx, next, 1000)
	require.NoError(t, err)
	require.Equal(t, len(res), 950)
	require.Equal(t, "", next)

}

func getMockModule() *storage.Version {
	return &storage.Version{
		Info: []byte("123"),
		Mod:  []byte("456"),
		Zip:  ioutil.NopCloser(bytes.NewReader([]byte("789"))),
	}
}
