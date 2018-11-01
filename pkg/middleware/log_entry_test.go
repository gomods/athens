package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/log"
)

func TestLogContext(t *testing.T) {
	h := func(c buffalo.Context) error {
		e := log.EntryFromContext(c, &log.Logger{})
		e.Infof("test")
		return nil
	}

	a := buffalo.New(buffalo.Options{})
	a.GET("/test", h)

	var buf bytes.Buffer
	lggr := log.New("", logrus.DebugLevel)
	lggr.Out = &buf

	a.Use(LogContextMiddleware(lggr))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test", nil)
	a.ServeHTTP(w, r)

	expected := `"http-method":"GET","http-path":"/test/","http-url":"/test/"`
	assert.True(t, strings.Contains(buf.String(), expected), fmt.Sprintf("%s should contain: %s", buf.String(), expected))
}
