package middleware

import (
	"context"
	"net/http"

	"github.com/gomods/athens/pkg/requestid"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// WithRequestID ensures a request id is in the request context.
// It prefers the trace ID extracted from incoming trace headers
// (traceparent/b3) via the global OTel propagator, then falls back
// to the Athens-Request-ID header, and finally generates a UUID.
func WithRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestID string
		// Derive a context from the request context but clear any existing
		// SpanContext so extraction reflects only incoming headers.
		cleanCtx := trace.ContextWithSpanContext(r.Context(), trace.SpanContext{})
		if sc := extractRemoteSpanContext(cleanCtx, r.Header); sc.HasTraceID() && sc.IsRemote() {
			// Use the extracted trace id only if it represents a remote
			// context (i.e. came from incoming headers). `IsRemote()`
			// indicates the propagator created this SpanContext from
			// incoming headers rather than it being a locally-created
			// context (for example from otelhttp).
			requestID = sc.TraceID().String()
		} else if id := r.Header.Get(requestid.HeaderKey); id != "" {
			requestID = id
		} else {
			requestID = uuid.New().String()
		}
		ctx := requestid.SetInContext(r.Context(), requestID)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

// extractRemoteSpanContext uses the global OTel propagator to extract
// a span context from the request headers. Returns an empty SpanContext
// if no valid trace headers are present.
func extractRemoteSpanContext(ctx context.Context, headers http.Header) trace.SpanContext {
	// Extract into the provided context (we pass context.Background()
	// at call sites) so we don't pick up any locally-created span
	// context already present on the request context (for example
	// from otelhttp). This ensures the extracted SpanContext
	// reflects only incoming headers.
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
	return trace.SpanContextFromContext(ctx)
}
