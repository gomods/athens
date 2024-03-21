package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/log"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
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
	lggr := log.New("", logrus.DebugLevel, "")
	lggr.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	lggr.Out = &buf

	r.Use(LogEntryMiddleware(lggr))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expected := `{"http-method":"GET","http-path":"/test","level":"info","msg":"test","request-id":""}`
	assert.True(t, strings.Contains(buf.String(), expected), fmt.Sprintf("%s should contain: %s", buf.String(), expected))
}
