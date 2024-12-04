package middleware

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLogContext(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		e := log.EntryFromContext(r.Context())
		e.Infof("test")
	}

	r := mux.NewRouter()
	r.HandleFunc("/test", h)

	var buf bytes.Buffer
	lggr := log.New("", slog.LevelDebug, "")
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(&buf, &opts)
	lggr.Logger = slog.New(handler)

	r.Use(LogEntryMiddleware(lggr))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expected := `{"http-method":"GET","http-path":"/test","level":"info","msg":"test","request-id":""}`
	assert.True(t, strings.Contains(buf.String(), expected), fmt.Sprintf("%s should contain: %s", buf.String(), expected))
}
