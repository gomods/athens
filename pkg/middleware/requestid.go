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
		if sc := extractRemoteSpanContext(r.Context(), r.Header); sc.HasTraceID() {
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
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(headers))
	return trace.SpanContextFromContext(ctx)
}
