package download

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestOfflineGoClient(t *testing.T) {
	const tag = "v1.0.0"
	const pseudo = "v1.0.1-0.20261002000000-abcdef123456"
	for _, tc := range []struct {
		name         string
		networkMode  string
		downloadMode mode.Mode
		versions     []string
		exclude      bool
		retract      bool
		want         string
		build        bool
	}{
		{"tagged only", Offline, mode.Sync, []string{tag}, false, false, tag, false},
		{"release preferred over commit and prerelease", Offline, mode.Sync, []string{tag, pseudo, "v1.1.0-rc1"}, false, false, tag, false},
		{"pseudo only", Offline, mode.Sync, []string{pseudo}, false, false, pseudo, false},
		{"excluded release", Offline, mode.Sync, []string{tag, pseudo}, true, false, pseudo, true},
		{"retracted release", Offline, mode.Sync, []string{tag, pseudo}, false, true, pseudo, false},
		{"none with excluded release", Strict, mode.None, []string{tag, pseudo}, true, false, pseudo, false},
		{"none with fallback policy", Fallback, mode.None, []string{tag, pseudo}, true, false, pseudo, false},
		{"fallback upstream outage", Fallback, mode.Sync, []string{tag, pseudo}, true, false, pseudo, false},
		{"missing module", Offline, mode.None, nil, false, false, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := mem.NewStorage()
			require.NoError(t, err)
			for _, version := range tc.versions {
				modFile := []byte("module " + cachedTestModule + "\n\ngo 1.24\n")
				if tc.retract && version == tag {
					modFile = append(modFile, []byte("\nretract "+tag+"\n")...)
				}
				saveCachedTestVersion(t, backend, version, modFile, cachedTestArchive(t, version, modFile))
			}
			lister := &cachedLatestLister{err: errors.E("upstream", "upstream unavailable", errors.KindGatewayTimeout)}
			df := &mode.DownloadFile{Mode: tc.downloadMode}
			dp := New(&Opts{Storage: backend, Lister: lister, NetworkMode: tc.networkMode, DownloadFile: df})
			router := mux.NewRouter()
			RegisterHandlers(router, &HandlerOpts{Protocol: dp, Logger: log.NoOpLogger(), DownloadFile: df})
			server := httptest.NewServer(router)
			t.Cleanup(server.Close)
			root := t.TempDir()
			modFile := "module example.com/client\n\ngo 1.24\n"
			if tc.exclude {
				modFile += "\nexclude " + cachedTestModule + " " + tag + "\n"
			}
			require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte(modFile), 0o600))
			env := offlineGoClientEnv(root, server.URL)
			out, err := runOfflineGo(t, root, env, "list", "-m", "-json", cachedTestModule+"@latest")
			if tc.want == "" {
				require.Error(t, err)
				require.Contains(t, string(out), "no matching versions")
			} else {
				require.NoError(t, err, "%s", out)
				var got storage.RevInfo
				require.NoError(t, json.Unmarshal(out, &got))
				require.Equal(t, tc.want, got.Version)
			}
			if tc.build {
				mainFile := "package main\n\nimport \"" + cachedTestModule + "\"\n\nfunc main() { println(offline.Value) }\n"
				require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte(mainFile), 0o600))
				out, err = runOfflineGo(t, root, env, "get", cachedTestModule+"@latest")
				require.NoError(t, err, "%s", out)
				out, err = runOfflineGo(t, root, env, "build", "-o", filepath.Join(root, "client"), ".")
				require.NoError(t, err, "%s", out)
			}
			resp, err := server.Client().Get(server.URL + "/" + cachedTestModule + "/@latest")
			require.NoError(t, err)
			defer resp.Body.Close()
			if tc.want == "" {
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				var got storage.RevInfo
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
				if len(tc.versions) == 1 && tc.versions[0] == tag {
					require.Equal(t, tag, got.Version)
				} else {
					require.Equal(t, pseudo, got.Version, "direct endpoint supplies the cached commit fallback")
				}
			}
			if tc.networkMode == Fallback && tc.downloadMode != mode.None {
				require.Positive(t, lister.calls.Load())
			} else {
				require.Zero(t, lister.calls.Load(), "offline discovery must not call upstream")
			}
		})
	}
}

func cachedTestArchive(t *testing.T, version string, modFile []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for name, data := range map[string][]byte{
		"go.mod":     modFile,
		"offline.go": []byte("package offline\n\nconst Value = 42\n"),
	} {
		file, err := writer.Create(cachedTestModule + "@" + version + "/" + name)
		require.NoError(t, err)
		_, err = file.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	return buf.Bytes()
}

func offlineGoClientEnv(root, proxy string) []string {
	return append(os.Environ(),
		"GOPROXY="+proxy,
		"GOSUMDB=off",
		"GOPRIVATE=", "GONOPROXY=", "GONOSUMDB=", "GOVCS=*:off",
		"GOENV=off", "GOWORK=off", "GOFLAGS=-modcacherw", "GO111MODULE=on", "GOTOOLCHAIN=local",
		"GOPATH="+filepath.Join(root, "gopath"),
		"GOMODCACHE="+filepath.Join(root, "modcache"),
	)
}

func runOfflineGo(t *testing.T, root string, env []string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	cmd.Env = env
	return cmd.CombinedOutput()
}
