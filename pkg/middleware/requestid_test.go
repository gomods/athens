package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/requestid"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func init() {
	// Set up composite propagator for tests (same as production).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(),
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

func TestWithRequestID(t *testing.T) {
	var givenRequestID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		givenRequestID = requestid.FromContext(r.Context())
	})

	t.Run("uses trace ID from W3C traceparent header", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Traceparent", "00-4bf92f3577b6a814af67ab2d6fc0f4e1-00f067aa0ba902b7-01")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID != "4bf92f3577b6a814af67ab2d6fc0f4e1" {
			t.Fatalf("expected trace id %q but got %q", "4bf92f3577b6a814af67ab2d6fc0f4e1", givenRequestID)
		}
	})

	t.Run("uses trace ID from B3 multi-header", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-B3-TraceId", "463ac35c9f6413ad48485a3953bb6124")
		req.Header.Set("X-B3-SpanId", "0020000000000001")
		req.Header.Set("X-B3-Sampled", "1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID != "463ac35c9f6413ad48485a3953bb6124" {
			t.Fatalf("expected trace id %q but got %q", "463ac35c9f6413ad48485a3953bb6124", givenRequestID)
		}
	})

	t.Run("uses trace ID from B3 single header", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("B3", "80f198ee56343ba864fe8b2a57d3eff7-e457b5a2e4d86bd1-1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID != "80f198ee56343ba864fe8b2a57d3eff7" {
			t.Fatalf("expected trace id %q but got %q", "80f198ee56343ba864fe8b2a57d3eff7", givenRequestID)
		}
	})

	t.Run("W3C takes priority over B3 when both present", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Traceparent", "00-aaaa1111bbbb2222cccc3333dddd4444-00f067aa0ba902b7-01")
		req.Header.Set("X-B3-TraceId", "1111222233334444aaaabbbbccccdddd")
		req.Header.Set("X-B3-SpanId", "0020000000000001")
		req.Header.Set("X-B3-Sampled", "1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID != "aaaa1111bbbb2222cccc3333dddd4444" {
			t.Fatalf("expected W3C trace id %q to win but got %q", "aaaa1111bbbb2222cccc3333dddd4444", givenRequestID)
		}
	})

	t.Run("falls back to Athens-Request-ID header when no trace headers", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(requestid.HeaderKey, "my-custom-id-123")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID != "my-custom-id-123" {
			t.Fatalf("expected %q but got %q", "my-custom-id-123", givenRequestID)
		}
	})

	t.Run("generates UUID when no trace headers and no Athens header", func(t *testing.T) {
		h := WithRequestID(handler)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if givenRequestID == "" {
			t.Fatal("expected a request id to be generated")
		}
		if _, err := uuid.Parse(givenRequestID); err != nil {
			t.Fatalf("expected a valid UUID but got %q: %v", givenRequestID, err)
		}
	})
}
