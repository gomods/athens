package azureblob

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage/compliance"
	"github.com/technosophos/moniker"
)

func TestBackend(t *testing.T) {
	backend := getStorage(t)
	defer func() {
		_, _ = backend.client.client.DeleteContainer(context.Background(), backend.client.containerName, nil)
	}()

	compliance.RunTests(t, backend, backend.clear)
}

func BenchmarkBackend(b *testing.B) {
	backend := getStorage(b)
	defer func() {
		_, _ = backend.client.client.DeleteContainer(context.Background(), backend.client.containerName, nil)
	}()

	compliance.RunBenchmarks(b, backend, backend.clear)
}

func (s *Storage) clear() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	pager := s.client.client.NewListBlobsFlatPager(s.client.containerName, nil)

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, blob := range resp.Segment.BlobItems {
			_, err := s.client.client.DeleteBlob(ctx, s.client.containerName, *blob.Name, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getStorage(t testing.TB) *Storage {
	t.Helper()

	containerName := randomContainerName(os.Getenv("GA_PULL_REQUEST"))

	cfg := getTestConfig(containerName)
	if cfg == nil {
		t.SkipNow()
	}

	s, err := New(cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.client.client.CreateContainer(context.Background(), s.client.containerName, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getTestConfig(containerName string) *config.AzureBlobConfig {
	key := os.Getenv("ATHENS_AZURE_ACCOUNT_KEY")
	resourceId := os.Getenv("ATHENS_AZURE_MANAGED_IDENTITY_RESOURCE_ID")
	credentialScope := os.Getenv("ATHENS_AZURE_CREDENTIAL_SCOPE")

	if key == "" && (resourceId == "" || credentialScope == "") {
		return nil
	}

	name := os.Getenv("ATHENS_AZURE_ACCOUNT_NAME")
	if name == "" {
		return nil
	}

	return &config.AzureBlobConfig{
		AccountName:               name,
		AccountKey:                key,
		ManagedIdentityResourceID: resourceId,
		CredentialScope:           credentialScope,
		ContainerName:             containerName,
	}
}

func randomContainerName(prefix string) string {
	// moniker is a cool library to produce mostly unique, human-readable names
	// see https://github.com/technosophos/moniker for more details
	namer := moniker.New()

	if prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, namer.NameSep(""))
	}

	return namer.NameSep("")
}
