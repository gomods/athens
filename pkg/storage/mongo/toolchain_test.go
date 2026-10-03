package mongo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReleaseObjectsLiteralKeys(t *testing.T) {
	backend := getStorage(t)
	backend.db = fmt.Sprintf("athens_toolchain_query_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		require.NoError(t, backend.client.Database(backend.db).Drop(context.Background()))
	})
	objects := &releaseObjects{backend}
	ctx := t.Context()
	body := []byte("unrelated archive")
	require.NoError(t, objects.Create(ctx, "unrelated", bytes.NewReader(body), int64(len(body))))

	read := func(t *testing.T, key string, want []byte) {
		t.Helper()
		reader, err := objects.Open(ctx, key)
		require.NoError(t, err)
		got, err := io.ReadAll(reader)
		require.NoError(t, reader.Close())
		require.NoError(t, err)
		require.Equal(t, want, got)
	}

	for _, key := range []string{`{"$ne":null}`, `{"$regex":".*"}`, `$where`, `'; return true; //`, ""} {
		t.Run(key, func(t *testing.T) {
			_, err := objects.Open(ctx, key)
			require.Equal(t, errors.KindNotFound, errors.Kind(err))
			err = objects.Delete(ctx, key)
			require.Equal(t, errors.KindNotFound, errors.Kind(err))
			read(t, "unrelated", body)

			literal := []byte("literal key: " + key)
			require.NoError(t, objects.Create(ctx, key, bytes.NewReader(literal), int64(len(literal))))
			read(t, key, literal)
			require.NoError(t, objects.Delete(ctx, key))
			_, err = objects.Open(ctx, key)
			require.Equal(t, errors.KindNotFound, errors.Kind(err))
			read(t, "unrelated", body)
		})
	}
}
