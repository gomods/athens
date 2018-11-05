package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
)

func TestUpstreamRedirect(t *testing.T) {
	h := func(c buffalo.Context) error { return errors.E("", errors.KindNotFound) }
	a := buffalo.New(buffalo.Options{})
	a.GET("/test", h)

	expected := "1.2.3.5"
	a.Use(NewUpstreamRedirectMiddleware(expected))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test", nil)
	a.ServeHTTP(w, r)

	actualStatus := w.Result().StatusCode
	if actualStatus != http.StatusSeeOther {
		t.Fatalf("Expected status see other got %v", actualStatus)
	}

	body, err := ioutil.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("Failed to read redirect uri '%v'", err)
	}

	actualRedirectURI := string(body)
	if !strings.Contains(actualRedirectURI, expected) {
		t.Fatalf("Expected redirect uri: '%v' got '%v'", expected, actualRedirectURI)
	}
}

func TestUpstreamRedirectNoUpstream(t *testing.T) {
	h := func(c buffalo.Context) error { return errors.E("", errors.KindNotFound) }
	a := buffalo.New(buffalo.Options{})
	a.GET("/test", h)

	expected := ""
	a.Use(NewUpstreamRedirectMiddleware(expected))

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test", nil)
	a.ServeHTTP(w, r)

	actualStatus := w.Result().StatusCode
	if actualStatus != http.StatusInternalServerError {
		t.Fatalf("Expected status see other got %v", actualStatus)
	}
}
