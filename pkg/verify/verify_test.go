package verify_test

import (
	"context"
	"testing"

	"github.com/gomods/athens/pkg/verify"
	"github.com/stretchr/testify/require"
)

func TestPoisoned(t *testing.T) {
	ctx := context.Background()
	b := newTestStore(t)
	zipBytes := writeZip(t, map[string]string{"example.com/m@v1.0.0/go.mod": "module example.com/m\n"})
	saveModule(t, ctx, b, "example.com/m", "v1.0.0", zipBytes)
	stored := hashOf(t, zipBytes)

	// Stored zip matches the canonical hash -> not poisoned.
	bad, err := verify.Poisoned(ctx, b, fakeOracle{map[string]string{"example.com/m@v1.0.0": stored}}, "example.com/m", "v1.0.0")
	require.NoError(t, err)
	require.False(t, bad)

	// Canonical differs -> poisoned.
	bad, err = verify.Poisoned(ctx, b, fakeOracle{map[string]string{"example.com/m@v1.0.0": "h1:DIFFERENT="}}, "example.com/m", "v1.0.0")
	require.NoError(t, err)
	require.True(t, bad)

	// Oracle can't find it -> error propagated (caller keeps the module).
	_, err = verify.Poisoned(ctx, b, fakeOracle{map[string]string{}}, "example.com/m", "v1.0.0")
	require.ErrorIs(t, err, verify.ErrNotInSumDB)
}
