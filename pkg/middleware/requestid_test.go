package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/requestid"
	"github.com/google/uuid"
)

func TestWithRequestID(t *testing.T) {
	var givenRequestID string
	h := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		givenRequestID = requestid.FromContext(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	expectedRequestID := uuid.New().String()
	req.Header.Set(requestid.HeaderKey, expectedRequestID)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if givenRequestID != expectedRequestID {
		t.Fatalf("expected request id to be %q but got %q", expectedRequestID, givenRequestID)
	}
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if givenRequestID == "" {
		t.Fatal("expected a request id to be created when a request id header is empty")
	}
}
