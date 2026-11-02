package s3

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveServerSideEncryption(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		algorithm    string
		keyID        string
		bucketKey    string
		cacheControl string
	}{
		{name: "bucket defaults"},
		{name: "SSE-S3", config: `ServerSideEncryption = "AES256"`, algorithm: "AES256"},
		{name: "SSE-KMS default key", config: `ServerSideEncryption = "aws:kms"`, algorithm: "aws:kms"},
		{
			name: "SSE-KMS key ID",
			config: `ServerSideEncryption = "aws:kms"
SSEKMSKeyID = "test-key-id"`,
			algorithm: "aws:kms", keyID: "test-key-id",
		},
		{
			name: "SSE-KMS key ARN and bucket key",
			config: `ServerSideEncryption = "aws:kms"
SSEKMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/test-key"
BucketKeyEnabled = true`,
			algorithm: "aws:kms", keyID: "arn:aws:kms:us-east-1:123456789012:key/test-key", bucketKey: "true",
		},
		{
			name: "SSE-KMS bucket key disabled",
			config: `ServerSideEncryption = "aws:kms"
BucketKeyEnabled = false`,
			algorithm: "aws:kms", bucketKey: "false",
		},
		{
			name:         "Cache-Control header",
			config:       `CacheControl = "public, max-age=31536000, immutable"`,
			cacheControl: "public, max-age=31536000, immutable",
		},
	}

	for _, tc := range tests {
		for _, multipart := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/multipart=%t", tc.name, multipart), func(t *testing.T) {
				var mu sync.Mutex
				headers := make(map[string]http.Header)
				bodies := make(map[string][]byte)
				parts := make(map[int][]byte)
				var initiated bool
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if !assert.NoError(t, err) {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					mu.Lock()
					defer mu.Unlock()
					query := r.URL.Query()
					switch {
					case r.Method == http.MethodPost && query.Has("uploads"):
						initiated = true
						headers[r.URL.Path] = r.Header.Clone()
						w.Header().Set("Content-Type", "application/xml")
						fmt.Fprint(w, `<InitiateMultipartUploadResult><UploadId>test-upload</UploadId></InitiateMultipartUploadResult>`)
					case r.Method == http.MethodPut && query.Has("partNumber"):
						n, err := strconv.Atoi(query.Get("partNumber"))
						assert.NoError(t, err)
						parts[n] = body
						w.Header().Set("ETag", fmt.Sprintf(`"part-%d"`, n))
					case r.Method == http.MethodPost && query.Has("uploadId"):
						for n := 1; n <= len(parts); n++ {
							bodies[r.URL.Path] = append(bodies[r.URL.Path], parts[n]...)
						}
						w.Header().Set("Content-Type", "application/xml")
						fmt.Fprint(w, `<CompleteMultipartUploadResult><ETag>"complete"</ETag></CompleteMultipartUploadResult>`)
					case r.Method == http.MethodPut:
						headers[r.URL.Path] = r.Header.Clone()
						bodies[r.URL.Path] = body
						w.Header().Set("ETag", `"object"`)
					default:
						t.Errorf("unexpected S3 request: %s %s", r.Method, r.URL)
						w.WriteHeader(http.StatusBadRequest)
					}
				}))
				defer server.Close()

				cfg := &config.S3Config{
					Bucket: "test-bucket", Region: "us-east-1", Endpoint: server.URL,
					ForcePathStyle: true, UseDefaultConfiguration: true,
				}
				_, err := toml.Decode(tc.config, cfg)
				require.NoError(t, err)
				backend, err := New(cfg, 30*time.Second, func(c *aws.Config) {
					c.Credentials = credentials.NewStaticCredentialsProvider("test-key", "test-secret", "")
				})
				require.NoError(t, err)
				mod := []byte("module example.com/module\n")
				info := []byte(`{"Version":"v1.0.0"}`)
				zip := []byte("zip contents")
				if multipart {
					// Exceed the SDK's default 16 MiB multipart threshold.
					zip = bytes.Repeat([]byte("z"), 17*1024*1024)
				}
				require.NoError(t, backend.Save(t.Context(), "example.com/module", "v1.0.0", mod, bytes.NewReader(zip), nil, info))

				mu.Lock()
				defer mu.Unlock()
				require.Len(t, headers, 3)
				assert.Equal(t, multipart, initiated)
				for ext, body := range map[string][]byte{"info": info, "mod": mod, "zip": zip} {
					path := "/test-bucket/" + config.PackageVersionedName("example.com/module", "v1.0.0", ext)
					h := headers[path]
					for name, want := range map[string]string{
						"X-Amz-Server-Side-Encryption":                    tc.algorithm,
						"X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id":     tc.keyID,
						"X-Amz-Server-Side-Encryption-Bucket-Key-Enabled": tc.bucketKey,
						"Cache-Control": tc.cacheControl,
					} {
						assert.Equal(t, want, h.Get(name), "%s: %s", ext, name)
						if want == "" {
							assert.NotContains(t, h, http.CanonicalHeaderKey(name), "%s: unset headers must be omitted", ext)
						}
					}
					assert.Equal(t, body, bodies[path], ext)
					assert.NotEmpty(t, h.Get("Content-Type"))
					assert.True(t, strings.HasPrefix(h.Get("Authorization"), "AWS4-HMAC-SHA256 "))
				}
			})
		}
	}
}

func TestCDNEndpointReadsAndChecks(t *testing.T) {
	const (
		mod  = "github.com/BurntSushi/toml"
		ver  = "v1.4.0"
		info = `{"Version":"v1.4.0"}`
		gom  = "module github.com/BurntSushi/toml\n"
		zip  = "zip-archive-bytes"
	)

	objects := map[string]string{
		"/" + config.PackageVersionedName(mod, ver, "info"): info,
		"/" + config.PackageVersionedName(mod, ver, "mod"):  gom,
		"/" + config.PackageVersionedName(mod, ver, "zip"):  zip,
	}

	var mu sync.Mutex
	var cdnRequests []string
	cdnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"), "CDN requests must not carry S3 Authorization headers")
		mu.Lock()
		cdnRequests = append(cdnRequests, r.Method+" "+r.URL.Path)
		mu.Unlock()

		if strings.Contains(r.URL.Path, "broken") {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		body, ok := objects[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	defer cdnServer.Close()

	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected S3 API call when CDNEndpoint is configured: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s3Server.Close()

	backend, err := New(&config.S3Config{
		Bucket:         "test-bucket",
		Region:         "us-east-1",
		Key:            "test-key",
		Secret:         "test-secret",
		Endpoint:       s3Server.URL,
		CDNEndpoint:    cdnServer.URL,
		ForcePathStyle: true,
	}, 30*time.Second)
	require.NoError(t, err)

	exists, err := backend.Exists(t.Context(), mod, ver)
	require.NoError(t, err)
	assert.True(t, exists)

	gotInfo, err := backend.Info(t.Context(), mod, ver)
	require.NoError(t, err)
	assert.Equal(t, []byte(info), gotInfo)

	gotMod, err := backend.GoMod(t.Context(), mod, ver)
	require.NoError(t, err)
	assert.Equal(t, []byte(gom), gotMod)

	zr, err := backend.Zip(t.Context(), mod, ver)
	require.NoError(t, err)
	assert.Equal(t, int64(len(zip)), zr.Size())
	gotZip, err := io.ReadAll(zr)
	require.NoError(t, err)
	require.NoError(t, zr.Close())
	assert.Equal(t, []byte(zip), gotZip)

	// Missing module returns false for Exists and KindNotFound for getters.
	exists, err = backend.Exists(t.Context(), mod, "v9.9.9")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = backend.Info(t.Context(), mod, "v9.9.9")
	require.True(t, errors.IsNotFoundErr(err), "expected NotFound error, got %v", err)

	_, err = backend.GoMod(t.Context(), mod, "v9.9.9")
	require.True(t, errors.IsNotFoundErr(err), "expected NotFound error, got %v", err)

	_, err = backend.Zip(t.Context(), mod, "v9.9.9")
	require.True(t, errors.IsNotFoundErr(err), "expected NotFound error, got %v", err)

	// Upstream 5xx propagates an error rather than a false NotFound.
	_, err = backend.Exists(t.Context(), "example.com/broken", "v1.0.0")
	require.Error(t, err)

	_, err = backend.Info(t.Context(), "example.com/broken", "v1.0.0")
	require.Error(t, err)
	assert.False(t, errors.IsNotFoundErr(err))

	mu.Lock()
	assert.NotEmpty(t, cdnRequests)
	mu.Unlock()
}

func TestParseCDNEndpointValidation(t *testing.T) {
	_, err := New(&config.S3Config{
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		CDNEndpoint: "ftp://invalid.example.com",
	}, 30*time.Second)
	require.Error(t, err)

	backend, err := New(&config.S3Config{
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		CDNEndpoint: "cdn.example.com/r2",
	}, 30*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example.com/r2/example.com/mod/@v/v1.0.0.info", backend.cdnObjectURL("example.com/mod/@v/v1.0.0.info"))
}
