package verify_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/verify"
	"github.com/stretchr/testify/require"
)

func TestSweep(t *testing.T) {
	ctx := context.Background()
	b := newTestStore(t)
	store := b.(verify.Store)

	goodZip := writeZip(t, map[string]string{"example.com/good@v1.0.0/go.mod": "module example.com/good\n"})
	badZip := writeZip(t, map[string]string{"example.com/bad@v1.0.0/go.mod": "module example.com/bad\n", "example.com/bad@v1.0.0/x.go": "package bad"})
	privZip := writeZip(t, map[string]string{"corp.internal/secret@v1.0.0/go.mod": "module corp.internal/secret\n"})

	saveModule(t, ctx, b, "example.com/good", "v1.0.0", goodZip)
	saveModule(t, ctx, b, "example.com/bad", "v1.0.0", badZip)
	saveModule(t, ctx, b, "corp.internal/secret", "v1.0.0", privZip)

	oracle := fakeOracle{hashes: map[string]string{
		"example.com/good@v1.0.0": hashOf(t, goodZip),  // matches stored
		"example.com/bad@v1.0.0":  "h1:CANONICAL-DIFFERENT=", // mismatches stored
		// corp.internal/secret intentionally absent (also skipped by NoSumPatterns)
	}}
	skip := func(mod string) bool { return strings.HasPrefix(mod, "corp.internal/") }

	// Report-only: detects the mismatch but deletes nothing.
	rep, err := verify.Sweep(ctx, store, oracle, skip, false, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 1, rep.Matched)
	require.Equal(t, 1, rep.Mismatched)
	require.Equal(t, 0, rep.Purged)
	require.Equal(t, 1, rep.SkippedPrivate)

	_, err = store.Zip(ctx, "example.com/bad", "v1.0.0")
	require.NoError(t, err, "report mode must not delete")

	// Purge: deletes only the mismatch.
	rep, err = verify.Sweep(ctx, store, oracle, skip, true, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 1, rep.Purged)

	_, err = store.Zip(ctx, "example.com/bad", "v1.0.0")
	require.Error(t, err, "purged version must be gone")
	_, err = store.Zip(ctx, "example.com/good", "v1.0.0")
	require.NoError(t, err, "matching version must remain")
	_, err = store.Zip(ctx, "corp.internal/secret", "v1.0.0")
	require.NoError(t, err, "private (skipped) version must remain")
}

func TestSweepSkippedUnknown(t *testing.T) {
	ctx := context.Background()
	b := newTestStore(t)
	store := b.(verify.Store)

	unknownZip := writeZip(t, map[string]string{"example.com/unknown@v1.0.0/go.mod": "module example.com/unknown\n"})
	saveModule(t, ctx, b, "example.com/unknown", "v1.0.0", unknownZip)

	// oracle has no entry for example.com/unknown@v1.0.0, so CanonicalHash
	// returns ErrNotInSumDB; the module is not private (skip never matches it).
	oracle := fakeOracle{hashes: map[string]string{}}
	skip := func(mod string) bool { return strings.HasPrefix(mod, "corp.internal/") }

	rep, err := verify.Sweep(ctx, store, oracle, skip, true, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 1, rep.SkippedUnknown)
	require.Equal(t, 0, rep.Matched)
	require.Equal(t, 0, rep.Mismatched)
	require.Equal(t, 0, rep.Purged)
	require.Equal(t, 0, rep.SkippedPrivate)
	require.Equal(t, 0, rep.Unverified)

	_, err = store.Zip(ctx, "example.com/unknown", "v1.0.0")
	require.NoError(t, err, "unknown (not in sumdb) version must not be deleted")
}
