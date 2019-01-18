package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	backend, closer := getStorage(t)
	defer closer()
	compliance.RunTests(t, backend, func() error {
		return nil // do nothing!
	})
}

func (s *ModuleStore) clear() error {
	fmt.Printf("Should clear---but didn't!\n")
	return nil
}

func BenchmarkBackend(b *testing.B) {
	backend, closer := getStorage(b)
	defer closer()
	compliance.RunBenchmarks(b, backend, func() error {
		return nil // do nothing!
	})
}

func getStorage(tb testing.TB) (*ModuleStore, func()) {

	s := httptest.NewServer(&httpServer{
		username: "some_username",
		password: "secret_password",
		files:    make(map[string][]byte),
	})

	backend, err := New(&config.HTTPConfig{
		BaseURL:  s.URL,
		Username: "some_username",
		Password: "secret_password",
	}, config.GetTimeoutDuration(300))
	require.NoError(tb, err)

	return backend, s.Close
}

type httpServer struct {
	username string
	password string
	files    map[string][]byte
	mu       sync.Mutex
}

func (hs *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// check Basic Auth, if any
	if hs.username != "" || hs.password != "" {
		if u, p, ok := r.BasicAuth(); !ok || u != hs.username || p != hs.password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	hs.mu.Lock()
	defer hs.mu.Unlock()

	if r.URL.Path == "/" {

		// special case: just be ok with this
		w.WriteHeader(http.StatusOK)
		return

	} else if strings.HasSuffix(r.URL.Path, "/@v/") {

		// you can only GET a directory listing
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var versions []string
		for k := range hs.files {
			if strings.HasPrefix(k, r.URL.Path) {
				versions = append(versions, strings.TrimPrefix(k, r.URL.Path))
			}
		}

		sort.Strings(versions)
		for _, v := range versions {
			fmt.Fprintf(w, `<a href="%s">%s</a>`, v, v)
		}

		return

	} else {

		// We're reading, writing, or deleting a file!

		switch r.Method {
		case http.MethodHead:
			if data, ok := hs.files[r.URL.Path]; ok {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		case http.MethodGet:
			if data, ok := hs.files[r.URL.Path]; ok {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
				w.Write(data)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		case http.MethodPut:
			if _, ok := hs.files[r.URL.Path]; ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data, _ := ioutil.ReadAll(r.Body)
			hs.files[r.URL.Path] = data
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			if _, ok := hs.files[r.URL.Path]; !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			delete(hs.files, r.URL.Path)
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	}

}
