package middleware

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gomods/athens/pkg/errors"
	"github.com/markbates/willie"
)

func TestUpstreamRedirect(t *testing.T) {
	h := func(c buffalo.Context) error {
		return c.Error(errors.KindNotFound, nil)
	}
	expected := "1.2.3.5"

	a := buffalo.New(buffalo.Options{})
	a.Use(NewUpstreamRedirectMiddleware(expected))
	a.GET("/test", h)

	w := willie.New(a)
	r := w.Request("/test").Get()

	actualStatus := r.Code
	expectedStatus := http.StatusSeeOther
	if actualStatus != expectedStatus {
		t.Fatalf("Expected status '%v' got '%v'", expectedStatus, actualStatus)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Failed to read redirect uri '%v'", err)
	}

	actualRedirectURI := string(body)
	if !strings.Contains(actualRedirectURI, expected) {
		t.Fatalf("Expected redirect uri: '%v' got '%v'", expected, actualRedirectURI)
	}
}

func TestUpstreamRedirectNoUpstream(t *testing.T) {
	h := func(c buffalo.Context) error {
		return c.Render(errors.KindNotFound, nil)
	}
	a := buffalo.New(buffalo.Options{})
	a.Use(NewUpstreamRedirectMiddleware(""))
	a.GET("/test", h)

	w := willie.New(a)
	r := w.Request("/test").Get()

	actualStatus := r.Code
	expectedStatus := http.StatusNotFound
	if actualStatus != expectedStatus {
		t.Fatalf("Expected status '%v' got '%v'", expectedStatus, actualStatus)
	}
}
