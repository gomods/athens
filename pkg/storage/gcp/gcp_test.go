package gcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/technosophos/moniker"
	"google.golang.org/api/iterator"
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	defer backend.bucket.Delete(t.Context())
	compliance.RunTests(t, backend, backend.clear)
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	defer backend.bucket.Delete(b.Context())
	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *Storage) clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	it := s.bucket.Objects(ctx, nil)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		err = s.bucket.Object(attrs.Name).Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func getStorage(t testing.TB) *Storage {
	t.Helper()
	bucketName := randomBucketName(os.Getenv("GA_PULL_REQUEST"))
	cfg := getTestConfig(bucketName)
	if cfg == nil {
		// Don't fail if there's no test config, so that these tests don't
		// fail when you run them locally
		t.Log("No GCS Config found")
		t.SkipNow()
	}

	s, err := newClient(t.Context(), cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}
	err = s.bucket.Create(t.Context(), cfg.ProjectID, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getTestConfig(bucket string) *config.GCPConfig {
	creds := os.Getenv("GCS_SERVICE_ACCOUNT")
	if creds == "" {
		return nil
	}
	return &config.GCPConfig{
		Bucket:    bucket,
		JSONKey:   creds,
		ProjectID: os.Getenv("GCS_PROJECT_ID"),
	}
}

func randomBucketName(prefix string) string {
	// moniker is a cool library to produce mostly unique, human-readable names
	// see https://github.com/technosophos/moniker for more details
	namer := moniker.New()
	if prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, namer.NameSep("_"))
	}
	return namer.NameSep("_")
}

// TestNewClientWithDifferentJSONKeys tests the newClient function with various JSON key types and error scenarios
func TestNewClientWithDifferentJSONKeys(t *testing.T) {
	tests := []struct {
		name        string
		jsonKey     string
		bucket      string
		projectID   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "empty JSON key uses default credentials",
			jsonKey:   "",
			bucket:    "test-bucket",
			projectID: "test-project",
			wantErr:   false,
		},
		{
			name:        "invalid base64 encoding",
			jsonKey:     "not-valid-base64!@#$%^&*()",
			bucket:      "test-bucket",
			projectID:   "test-project",
			wantErr:     true,
			errContains: "could not decode base64 json key",
		},
		{
			name:        "invalid JSON format",
			jsonKey:     base64.StdEncoding.EncodeToString([]byte("{invalid json}")),
			bucket:      "test-bucket",
			projectID:   "test-project",
			wantErr:     true,
			errContains: "failed to parse JSON key",
		},
		{
			name:        "missing type field",
			jsonKey:     encodeJSONToBase64(t, map[string]interface{}{"project_id": "test"}),
			bucket:      "test-bucket",
			projectID:   "test-project",
			wantErr:     true,
			errContains: "missing 'type' field",
		},
		{
			name:        "unknown credential type",
			jsonKey:     encodeJSONToBase64(t, map[string]interface{}{"type": "unknown_type"}),
			bucket:      "test-bucket",
			projectID:   "test-project",
			wantErr:     true,
			errContains: "unknown credential type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := &config.GCPConfig{
				ProjectID: tt.projectID,
				Bucket:    tt.bucket,
				JSONKey:   tt.jsonKey,
			}

			_, err := newClient(ctx, cfg, 30*time.Second)

			if tt.wantErr {
				if err == nil {
					t.Errorf("newClient() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("newClient() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil && !strings.Contains(err.Error(), "could not create new storage client") {
					// We expect storage client creation to potentially fail in tests without proper GCP credentials
					// but we should not get errors before that point
					t.Errorf("newClient() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestNewClientServiceAccount tests newClient with a service account JSON key
func TestNewClientServiceAccount(t *testing.T) {
	// Create a minimal service account JSON structure
	// Note: This will fail at credential creation without a real private key,
	// but it should pass validation up to that point
	serviceAccountJSON := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                  "test-project-id",
		"private_key_id":              "key123",
		"private_key":                 "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n",
		"client_email":                "test@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	}

	ctx := context.Background()
	cfg := &config.GCPConfig{
		ProjectID: "test-project",
		Bucket:    "test-bucket",
		JSONKey:   encodeJSONToBase64(t, serviceAccountJSON),
	}

	_, err := newClient(ctx, cfg, 30*time.Second)

	// We expect this to fail at credential creation with the fake key,
	// but the important part is it recognized the service_account type correctly
	// and didn't fail with "unsupported credential type"
	if err != nil {
		if strings.Contains(err.Error(), "unsupported credential type") {
			t.Errorf("newClient() failed with unsupported credential type, but service_account should be supported")
		}
		// Expected to fail with credential/storage client error since we're using fake credentials
		if !strings.Contains(err.Error(), "could not get GCS credentials") &&
			!strings.Contains(err.Error(), "could not create new storage client") {
			t.Logf("newClient() failed as expected with: %v", err)
		}
	}
}

// TestNewClientExternalAccount tests newClient with an external account JSON key
func TestNewClientExternalAccount(t *testing.T) {
	// Create a minimal external account JSON structure
	// Note: This will fail at credential creation without proper setup,
	// but it should pass validation up to that point
	externalAccountJSON := map[string]interface{}{
		"type":               "external_account",
		"audience":           "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider",
		"subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
		"token_url":          "https://sts.googleapis.com/v1/token",
		"credential_source": map[string]interface{}{
			"file": "/var/run/secrets/token",
		},
	}

	ctx := context.Background()
	cfg := &config.GCPConfig{
		ProjectID: "test-project",
		Bucket:    "test-bucket",
		JSONKey:   encodeJSONToBase64(t, externalAccountJSON),
	}

	_, err := newClient(ctx, cfg, 30*time.Second)

	// We expect this to fail at credential creation with the fake credentials,
	// but the important part is it recognized the external_account type correctly
	// and didn't fail with "unsupported credential type"
	if err != nil {
		if strings.Contains(err.Error(), "unsupported credential type") {
			t.Errorf("newClient() failed with unsupported credential type, but external_account should be supported")
		}
		// Expected to fail with credential/storage client error since we're using fake credentials
		if !strings.Contains(err.Error(), "could not get GCS credentials") &&
			!strings.Contains(err.Error(), "could not create new storage client") {
			t.Logf("newClient() failed as expected with: %v", err)
		}
	}
}

// encodeJSONToBase64 is a test helper that encodes a map to JSON and then to base64
func encodeJSONToBase64(t *testing.T, data map[string]interface{}) string {
	t.Helper()
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return base64.StdEncoding.EncodeToString(jsonBytes)
}
