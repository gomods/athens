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

func TestAppReturnsCleanupWithExporters(t *testing.T) {
	l := log.NoOpLogger()
	c, err := config.Load("")
	require.NoError(t, err)
	// Exercise the real exporter registration path: OTLP traces (the gRPC
	// exporter dials lazily, so no collector is needed) and Prometheus metrics.
	c.TraceExporter = "otlp"
	c.StatsExporter = "prometheus"

	handler, cleanup, err := App(l, c)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, cleanup)

	// cleanup shuts down the trace and metric providers and must be safe to call.
	assert.NotPanics(t, cleanup)

	// Calling cleanup a second time should also be safe (idempotency).
	assert.NotPanics(t, cleanup)
}
