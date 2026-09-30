package paths

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestMatchesPattern(t *testing.T) {
	type args struct {
		pattern string
		name    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "standard match",
			args: args{
				pattern: "example.com/*",
				name:    "example.com/athens",
			},
			want: true,
		},
		{
			name: "mutiple depth match",
			args: args{
				pattern: "example.com/*",
				name:    "example.com/athens/pkg",
			},
			want: true,
		},
		{
			name: "subdomain match",
			args: args{
				pattern: "*.example.com/*",
				name:    "go.example.com/athens/pkg",
			},
			want: true,
		},
		{
			name: "subdirectory exact match",
			args: args{
				pattern: "*.example.com/mod",
				name:    "go.example.com/mod/example",
			},
			want: true,
		},
		{
			name: "subdirectory mismatch",
			args: args{
				pattern: "*.example.com/mod",
				name:    "go.example.com/pkg/example",
			},
			want: false,
		},
		{
			name: "shorter name mismatch",
			args: args{
				pattern: "*.example.com/mod/pkg",
				name:    "go.example.com/pkg",
			},
			want: false,
		},
		{
			name: "no subdirectory mismatch",
			args: args{
				pattern: "*.example.com/mod/pkg",
				name:    "go.example.com/pkg",
			},
			want: false,
		},
		{
			name: "bad pattern",
			args: args{
				pattern: "[]a]",
				name:    "go.example.com/pkg",
			},
			want: false,
		},
		{
			name: "matches everything",
			args: args{
				pattern: "*",
				name:    "github.com/gomods/athen",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesPattern(tt.args.pattern, tt.args.name)
			if got != tt.want {
				t.Errorf("MatchGlobPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProtocolPath(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		route  string
		path   string
		want   string
	}{
		{
			name:  "no prefix",
			route: "/{module:.+}/@v/{version}.info",
			path:  "/example.com/mod/@v/v1.0.0.info",
			want:  "/example.com/mod/@v/v1.0.0.info",
		},
		{
			name:   "prefix",
			prefix: "/prefix",
			route:  "/{module:.+}/@v/{version}.info",
			path:   "/prefix/example.com/mod/@v/v1.0.0.info",
			want:   "/example.com/mod/@v/v1.0.0.info",
		},
		{
			name:   "prefix with trailing slash",
			prefix: "/prefix/",
			route:  "/{module:.+}/@latest",
			path:   "/prefix/example.com/mod/@latest",
			want:   "/example.com/mod/@latest",
		},
		{
			name:   "prefix equal to module prefix",
			prefix: "/example.com",
			route:  "/{module:.+}/@v/list",
			path:   "/example.com/example.com/mod/@v/list",
			want:   "/example.com/mod/@v/list",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			r := mux.NewRouter()
			sr := r
			if tt.prefix != "" {
				sr = r.PathPrefix(tt.prefix).Subrouter()
			}
			sr.HandleFunc(tt.route, func(_ http.ResponseWriter, r *http.Request) {
				got = ProtocolPath(r)
			})
			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, tt.path, nil))
			if got != tt.want {
				t.Errorf("ProtocolPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkMatchesPattern(b *testing.B) {
	for i := 1; i < 5; i++ {
		target := "git.example.com" + strings.Repeat("/path", i) + "/pkg"
		b.Run(fmt.Sprintf("MatchPattern/%d", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if !MatchesPattern("*.example.com/*", target) {
					b.Error("mismatch")
				}
			}
		})
	}
}
