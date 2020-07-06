package stash

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/internal/testutil"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// TestEtcdSingleFlight will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestEtcdSingleFlight(t *testing.T) {
	strg, err := mem.NewStorage()
	if err != nil {
		t.Fatal(err)
	}
	ms := &mockEtcdStasher{strg: strg}
	endpoints := strings.Split(etcdTestConfig(t).Endpoints, ",")
	wrapper, err := WithEtcd(endpoints, storage.WithChecker(strg))
	if err != nil {
		t.Fatal(err)
	}
	s := wrapper(ms)

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
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

func etcdTestConfig(t *testing.T) *config.Etcd {
	t.Helper()
	if os.Getenv("SKIP_ETCD") != "" {
		t.SkipNow()
		return nil
	}

	c, err := config.Load("")
	require.NoError(t, err)
	cfg := c.SingleFlight.Etcd
	if os.Getenv("STATIC_PORTS") == "" {
		var eps []string
		for _, host := range []string{"etcd0", "etcd1", "etcd2"} {
			h, p := testutil.GetServicePort(t, host, 2379)
			eps = append(eps, fmt.Sprintf("http://%s:%d", h, p))
		}
		cfg.Endpoints = strings.Join(eps, ",")
	}
	return cfg
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
