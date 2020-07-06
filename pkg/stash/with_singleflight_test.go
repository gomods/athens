package stash

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSingleFlight will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestSingleFlight(t *testing.T) {
	checkerCount := 0
	checker := mockChecker(func(_ context.Context, mod, ver string) (bool, error) {
		t.Helper()
		assert.Equal(t, "mod", mod)
		assert.Equal(t, "ver", ver)
		checkerCount++
		if checkerCount > 1 {
			return true, nil
		}
		return false, nil
	})
	stashCount := 0
	ms := mockStasher(func(_ context.Context, mod, ver string) (string, error) {
		t.Helper()
		assert.Equal(t, "mod", mod)
		assert.Equal(t, "ver", ver)
		stashCount++
		time.Sleep(100 * time.Millisecond)
		return "newVer", nil
	})
	wrapper := WithSingleflight(checker)
	s := wrapper(ms)
	for i := 0; i < 5; i++ {
		_, err := s.Stash(context.Background(), "mod", "ver")
		require.NoError(t, err)
	}
	require.Equal(t, 1, stashCount)
	require.Equal(t, 5, checkerCount)
}

type mockChecker func(ctx context.Context, module, version string) (bool, error)

func (m mockChecker) Exists(ctx context.Context, module, version string) (bool, error) {
	return m(ctx, module, version)
}

type mockStasher func(ctx context.Context, mod, ver string) (string, error)

func (m mockStasher) Stash(ctx context.Context, mod string, ver string) (string, error) {
	return m(ctx, mod, ver)
}
