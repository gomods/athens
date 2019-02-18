package stash

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"golang.org/x/sync/errgroup"
)

// TestEtcdSingleFlight will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestEtcdSingleFlight(t *testing.T) {
	endpointsStr := os.Getenv("ETCD_TEST_ENDPOINTS")
	if len(endpointsStr) == 0 {
		t.SkipNow()
	}
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockEtcdStasher{strg: strg}
	endpoints := strings.Split(endpointsStr, ",")
	wrapper, err := WithEtcd(endpoints, strg)
	if err != nil {
		t.Fatal(err)
	}
	s := wrapper(ms)

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			_, err := s.Stash(context.Background(), "mod", "ver")
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

// mockEtcdStasher is like mockStasher
// but leverages in memory storage
// so that etcd can determine
// whether to call the underlying stasher or not.
type mockEtcdStasher struct {
	strg storage.Backend
	mu   sync.Mutex
	num  int
}

func (ms *mockEtcdStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
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
