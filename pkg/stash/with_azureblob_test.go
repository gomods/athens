package stash

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/azureblob"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// TestWithAzureBlob requires a real AzureBlob backend implementation
// and it will ensure that saving to modules at the same time
// is done synchronously so that only the first module gets saved.
func TestWithAzureBlob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	const (
		mod = "stashmod"
		ver = "v1.0.0"
	)
	cfg := getAzureTestConfig()
	strg := getAzureStorage(t)
	strg.Delete(ctx, mod, ver)
	defer strg.Delete(ctx, mod, ver)

	// sanity check
	_, err := strg.GoMod(ctx, mod, ver)
	if !errors.Is(err, errors.KindNotFound) {
		t.Fatalf("expected the stash bucket to return a NotFound error but got: %v", err)
	}

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		content := uuid.New().String()
		ms := &mockAzureBlobStasher{strg, content}
		wpr, err := WithAzureBlobLock(cfg, time.Second*10, strg)
		if err != nil {
			t.Fatal(err)
		}
		s := wpr(ms)
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			_, err := s.Stash(ctx, "stashmod", "v1.0.0")
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
	info, err := strg.Info(ctx, mod, ver)
	if err != nil {
		t.Fatal(err)
	}
	modContent, err := strg.GoMod(ctx, mod, ver)
	if err != nil {
		t.Fatal(err)
	}
	zip, err := strg.Zip(ctx, mod, ver)
	if err != nil {
		t.Fatal(err)
	}
	defer zip.Close()
	zipContent, err := ioutil.ReadAll(zip)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(info, modContent) {
		t.Fatalf("expected info and go.mod to be equal but info was {%v} and content was {%v}", string(info), string(modContent))
	}
	if !bytes.Equal(info, zipContent) {
		t.Fatalf("expected info and zip to be equal but info was {%v} and content was {%v}", string(info), string(zipContent))
	}
}

// mockAzureBlobStasher is like mockStasher
// but leverages in memory storage
// so that azure blob can determine
// whether to call the underlying stasher or not.
type mockAzureBlobStasher struct {
	strg    storage.Backend
	content string
}

func (ms *mockAzureBlobStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	err := ms.strg.Save(
		ctx,
		mod,
		ver,
		[]byte(ms.content),
		strings.NewReader(ms.content),
		[]byte(ms.content),
	)
	return "", err
}

func getAzureStorage(t *testing.T) *azureblob.Storage {
	t.Helper()
	cfg := getAzureTestConfig()
	if cfg == nil {
		t.SkipNow()
	}

	s, err := azureblob.New(cfg, config.GetTimeoutDuration(30))
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func getAzureTestConfig() *config.AzureBlobConfig {
	key := os.Getenv("ATHENS_AZURE_ACCOUNT_KEY")
	if key == "" {
		return nil
	}
	return &config.AzureBlobConfig{
		AccountName:   "athens_drone_azure_account",
		AccountKey:    key,
		ContainerName: "athens_drone_azure_container",
	}
}
