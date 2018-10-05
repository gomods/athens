package stash

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// TestSingleFlight will ensure that 5 concurrent requests will all get the first request's
// response. We can ensure that because only the first response does not return an error
// and therefore all 5 responses should have no error.
func TestSingleFlight(t *testing.T) {
	ms := &mockSFStasher{}
	s := WithSingleflight(ms)

	var eg errgroup.Group
	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			return s.Stash(context.Background(), "mod", "ver")
		})
	}

	err := eg.Wait()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		eg.Go(func() error {
			return s.Stash(context.Background(), "mod", "ver")
		})
	}
	err = eg.Wait()
	if err == nil {
		t.Fatal("expected second error to return")
	}
}

// mockSFStasher mocks a Stash request that
// will always return a different result after the
// first one. This way we can prove that a second
// request did not get a second result, but the first
// one, provided the request came in at the right time.
type mockSFStasher struct {
	mu  sync.Mutex
	num int
}

func (ms *mockSFStasher) Stash(ctx context.Context, mod, ver string) error {
	time.Sleep(time.Millisecond * 100) // allow for second requests to come in.
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.num == 0 {
		ms.num++
		return nil
	}
	return fmt.Errorf("second time error")
}
