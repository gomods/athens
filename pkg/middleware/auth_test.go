package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/auth"
)

func TestAuthMiddleware(t *testing.T) {
	var tests = []struct {
		name     string
		reqfunc  func(r *http.Request)
		wantok   bool
		wantauth auth.BasicAuth
	}{
		{
			name:    "no auth",
			reqfunc: func(r *http.Request) {},
		},
		{
			name: "with basic auth",
			reqfunc: func(r *http.Request) {
				r.SetBasicAuth("user", "pass")
			},
			wantok:   true,
			wantauth: auth.BasicAuth{User: "user", Password: "pass"},
		},
		{
			name: "only user",
			reqfunc: func(r *http.Request) {
				r.SetBasicAuth("justuser", "")
			},
			wantok:   true,
			wantauth: auth.BasicAuth{User: "justuser"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var (
				givenok   bool
				givenauth auth.BasicAuth
			)
			h := WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				givenauth.User, givenauth.Password, givenok = r.BasicAuth()
			}))

			r := httptest.NewRequest("GET", "/", nil)
			tc.reqfunc(r)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)

			if givenok != tc.wantok {
				t.Fatalf("expected basic auth existence to be %t but got %t", tc.wantok, givenok)
			}
			if givenauth != tc.wantauth {
				t.Fatalf("expected basic auth to be %+v but got %+v", tc.wantauth, givenauth)
			}
		})
	}
}
