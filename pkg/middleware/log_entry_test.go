package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
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

	buf := new(bytes.Buffer)
	lggr := log.New("", slog.LevelInfo, "")

	r.Use(LogEntryMiddleware(lggr))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	expectedFields := map[string]interface{}{
		"level":       "INFO",
		"msg":         "test",
		"http-method": "GET",
		"http-path":   "/test",
		"request-id":  "",
	}

	for k, v := range expectedFields {
		assert.Equal(t, v, logEntry[k], "Log entry should contain %s with value %v", k, v)
	}
}
