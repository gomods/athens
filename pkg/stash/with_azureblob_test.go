package stash

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/technosophos/moniker"
	"golang.org/x/sync/errgroup"
)

// TestWithAzureBlob requires a real AzureBlob backend implementation
// and it will ensure that saving to modules at the same time
// is done synchronously so that only the first module gets saved.
func TestWithAzureBlob(t *testing.T) {
	containerName := randomContainerName(os.Getenv("GA_PULL_REQUEST"))
	cfg := getAzureTestConfig(containerName)
	if cfg == nil {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockAzureBlobStasher{strg: strg}
	wpr, err := WithAzureBlobLock(cfg, time.Second*10, storage.WithChecker(strg))
	if err != nil {
		t.Fatal(err)
	}
	s := wpr(ms)

	var eg errgroup.Group
	for range 5 {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(t.Context(), time.Second*10)
			defer cancel()
			_, err := s.Stash(ctx, "mod", "ver")
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

// mockAzureBlobStasher is like mockStasher
// but leverages in memory storage
// so that azure blob can determine
// whether to call the underlying stasher or not.
type mockAzureBlobStasher struct {
	strg storage.Backend
	mu   sync.Mutex
	num  int
}

func (ms *mockAzureBlobStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	time.Sleep(time.Millisecond * 100) // allow for second requests to come in.
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.num == 0 {
		err := ms.strg.Save(
			ctx,
			mod,
			ver,
			[]byte("mod file"),
			strings.NewReader("zip file"),
			nil,
			[]byte("info file"),
		)
		if err != nil {
			return "", err
		}
		ms.num++
		return "", nil
	}
	return "", fmt.Errorf("second time error")
}

func getAzureTestConfig(containerName string) *config.AzureBlobConfig {
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

// azureBlockingStasher blocks in Stash until release is closed.
type azureBlockingStasher struct {
	started chan struct{}
	release chan struct{}
}

func newAzureBlockingStasher() *azureBlockingStasher {
	return &azureBlockingStasher{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (b *azureBlockingStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	close(b.started)
	<-b.release
	return "v", nil
}

// waitForGoroutines fails the test unless the number of live goroutines
// drops back to at most want, which means waitForStash's goroutine exited.
func waitForGoroutines(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("stasher goroutine leaked: have %d goroutines, want <= %d", runtime.NumGoroutine(), want)
}

func TestWaitForStashNoGoroutineLeakOnContextCancel(t *testing.T) {
	before := runtime.NumGoroutine()
	st := newAzureBlockingStasher()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		_, err := waitForStash(ctx, "test", st, "mod", "v1.0.0", time.Hour, func(context.Context) error { return nil })
		errCh <- err
	}()

	<-st.started
	cancel()
	if err := <-errCh; err == nil {
		t.Fatal("expected error after context cancellation")
	}

	// waitForStash has returned and nobody receives from its result channel
	// anymore. The stasher can now finish; its goroutine must still exit.
	close(st.release)
	waitForGoroutines(t, before)
}

func TestWaitForStashNoGoroutineLeakOnRenewFailure(t *testing.T) {
	before := runtime.NumGoroutine()
	st := newAzureBlockingStasher()
	renewErr := errors.New("renew failed")

	_, err := waitForStash(context.Background(), "test", st, "mod", "v1.0.0", time.Millisecond,
		func(context.Context) error { return renewErr })
	if err == nil || !strings.Contains(err.Error(), renewErr.Error()) {
		t.Fatalf("expected renew error, got %v", err)
	}

	close(st.release)
	waitForGoroutines(t, before)
}

func TestWaitForStashReturnsResult(t *testing.T) {
	st := newAzureBlockingStasher()
	close(st.release)
	v, err := waitForStash(context.Background(), "test", st, "mod", "v1.0.0", time.Hour, func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if v != "v" {
		t.Fatalf("expected v, got %q", v)
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
