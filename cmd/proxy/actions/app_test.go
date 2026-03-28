package actions

import (
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppReturnsCleanup(t *testing.T) {
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)

	handler, cleanup, err := App(l, c)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, cleanup)

	// cleanup should be safe to call without panic.
	assert.NotPanics(t, cleanup)
}

func TestAppReturnsCleanupWithDatadogExporter(t *testing.T) {
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)
	c.TraceExporter = "datadog"
	c.StatsExporter = "datadog"

	handler, cleanup, err := App(l, c)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, cleanup)

	// Datadog's RegisterExporter returns ex.Stop (not ex.Flush).
	// Before the fix, defer in App() called Stop() on return — before the
	// server even started — killing the background flush loop.
	// Now cleanup is returned to the caller and should be safe to invoke.
	assert.NotPanics(t, cleanup)

	// Calling cleanup a second time should also be safe (idempotency).
	assert.NotPanics(t, cleanup)
}
