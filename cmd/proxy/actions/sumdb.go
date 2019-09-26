package actions

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

func sumdbPoxy(url *url.URL, nosumPatterns []string) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(url)
	rp.Director = func(req *http.Request) {
		req.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
	}
	if len(nosumPatterns) > 0 {
		return noSumWrapper(rp, url.Host, nosumPatterns)
	}
	return rp
}

func noSumWrapper(h http.Handler, host string, patterns []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/lookup/") {
			for _, p := range patterns {
				if matchesPattern(p, r.URL.Path[len("/lookup/"):]) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}

// matchesPattern is adopted from
// https://github.com/golang/go/blob/a11644a26557ea436d456f005f39f4e01902bafe/src/cmd/go/internal/str/path.go#L58
// this function matches based on path prefixes and
// tries to keep the same behavior as GONOSUMDB and friends
func matchesPattern(pattern, target string) bool {
	n := strings.Count(pattern, "/")
	prefix := target
	for i := 0; i < len(target); i++ {
		if target[i] == '/' {
			if n == 0 {
				prefix = target[:i]
				break
			}
			n--
		}
	}
	if n > 0 {
		return false
	}
	matched, _ := path.Match(pattern, prefix)
	if matched {
		return true
	}
	return false
}
