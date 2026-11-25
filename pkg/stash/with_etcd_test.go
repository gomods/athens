package stash

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/client/v3/concurrency"
	etcdembed "go.etcd.io/etcd/server/v3/embed"
)

type testContextKey struct{}

type testError string

func (e testError) Error() string { return string(e) }

type mockChecker struct {
	t                       *testing.T
	e                       *etcdembed.Etcd
	wantModule, wantVersion string

	exists bool
	err    error
}

func (c *mockChecker) Exists(ctx context.Context, module, version string) (bool, error) {
	assert.NotNil(c.t, ctx.Value(testContextKey{}))
	assert.Equal(c.t, c.wantModule, module)
	assert.Equal(c.t, c.wantVersion, version)

	res, err := c.e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
	require.NoError(c.t, err)

	assert.Len(c.t, res.Leases, 1)

	return c.exists, c.err
}

type mockStasher struct {
	t                       *testing.T
	wantModule, wantVersion string

	newVersion string
	err        error
}

func (s *mockStasher) Stash(ctx context.Context, module, version string) (string, error) {
	assert.NotNil(s.t, ctx.Value(testContextKey{}))
	assert.Equal(s.t, s.wantModule, module)
	assert.Equal(s.t, s.wantVersion, version)
	return s.newVersion, s.err
}

func startClosableEtcdServer(t *testing.T) (endpoint string, e *etcdembed.Etcd, closeServer func()) {
	c := etcdembed.NewConfig()
	c.Dir = t.TempDir()
	c.ListenPeerUrls = []url.URL{{Host: "localhost:0"}}
	c.ListenClientUrls = []url.URL{{Host: "localhost:0"}}

	e, err := etcdembed.StartEtcd(c)
	if err != nil {
		t.Fatal(err)
	}
	var once sync.Once
	closeServer = func() {
		once.Do(func() { e.Close() })
	}
	t.Cleanup(closeServer)

	deadline, ok := t.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(time.Until(deadline)):
		t.Fatal("etcd server startup duration exceeded deadline")
	}
	return e.Clients[0].Addr().String(), e, closeServer
}

func startEtcdServer(t *testing.T) (endpoint string, e *etcdembed.Etcd) {
	endpoint, e, _ = startClosableEtcdServer(t)
	return endpoint, e
}

func TestWithEtcd(t *testing.T) {
	endpoint, _ := startEtcdServer(t)

	var someChecker mockChecker
	w, err := WithEtcd([]string{endpoint}, &someChecker)
	require.NoError(t, err)

	var someStasher mockStasher
	etcdSingleflight, ok := w(&someStasher).(*etcd)
	require.True(t, ok)

	assert.Equal(t, etcdSingleflight.checker, &someChecker)
	assert.Equal(t, etcdSingleflight.stasher, &someStasher)
}

func TestEtcdStash(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		const (
			someModule  = "some module"
			someVersion = "some version"
		)

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
			exists:      true,
		})
		require.NoError(t, err)

		st := w(&mockStasher{
			t:           t,
			wantModule:  someModule,
			wantVersion: someVersion,
		})

		newVersion, err := st.Stash(ctx, someModule, someVersion)
		require.NoError(t, err)
		assert.Equal(t, someVersion, newVersion)

		res, err := e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		require.Len(t, res.Leases, 1)
		firstLeaseID := res.Leases[0].ID

		newVersion, err = st.Stash(ctx, someModule, someVersion)
		require.NoError(t, err)
		assert.Equal(t, someVersion, newVersion)

		res, err = e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		require.Len(t, res.Leases, 1)
		assert.Equal(t, firstLeaseID, res.Leases[0].ID)
	})

	t.Run("module does not exist", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		const (
			someModule  = "some module"
			someVersion = "some version"

			someNewVersion = "some new version"
		)

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
		})
		require.NoError(t, err)

		newVersion, err := w(&mockStasher{
			t:           t,
			wantModule:  someModule,
			wantVersion: someVersion,
			newVersion:  someNewVersion,
		}).Stash(ctx, someModule, someVersion)
		require.NoError(t, err)

		assert.Equal(t, someNewVersion, newVersion)
	})

	t.Run("checker error", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		const (
			someModule  = "some module"
			someVersion = "some version"

			someError testError = "some error"
		)

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
			err:         someError,
		})
		require.NoError(t, err)

		newVersion, err := w(&mockStasher{}).Stash(ctx, someModule, someVersion)
		assert.ErrorIs(t, err, someError)
		assert.Empty(t, newVersion)
	})

	t.Run("stasher error", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		const (
			someModule  = "some module"
			someVersion = "some version"

			someError testError = "some error"
		)

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
		})
		require.NoError(t, err)

		newVersion, err := w(&mockStasher{
			t:           t,
			wantModule:  someModule,
			wantVersion: someVersion,
			err:         someError,
		}).Stash(ctx, someModule, someVersion)
		assert.ErrorIs(t, err, someError)
		assert.Empty(t, newVersion)
	})

	t.Run("canceled context does not create session", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		w, err := WithEtcd([]string{endpoint}, &mockChecker{t: t, e: e})
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err = w(&mockStasher{}).Stash(ctx, "example.com/mod", "v1.0.0")
		require.ErrorIs(t, err, context.Canceled)

		res, err := e.Server.LeaseLeases(t.Context(), &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		assert.Empty(t, res.Leases)
	})

	t.Run("concurrent same-key calls do not overlap on shared session", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)
		const (
			someModule  = "example.com/mod"
			someVersion = "v1.0.0"
		)

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
		})
		require.NoError(t, err)

		var active, maxActive, calls atomic.Int32
		st := w(stasherFunc(func(_ context.Context, mod, ver string) (string, error) {
			calls.Add(1)
			cur := active.Add(1)
			for {
				prev := maxActive.Load()
				if cur <= prev || maxActive.CompareAndSwap(prev, cur) {
					break
				}
			}
			time.Sleep(30 * time.Millisecond)
			active.Add(-1)
			return ver, nil
		}))

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, stashErr := st.Stash(ctx, someModule, someVersion)
				errs <- stashErr
			}()
		}
		wg.Wait()
		close(errs)
		for stashErr := range errs {
			require.NoError(t, stashErr)
		}

		assert.Equal(t, int32(2), calls.Load())
		assert.Equal(t, int32(1), maxActive.Load())

		res, err := e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		assert.Len(t, res.Leases, 1)
	})

	t.Run("revoked session is invalidated and replaced", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)
		const (
			someModule  = "example.com/mod"
			someVersion = "v1.0.0"
		)

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
			exists:      true,
		})
		require.NoError(t, err)

		st := w(&mockStasher{})
		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})

		ver, err := st.Stash(ctx, someModule, someVersion)
		require.NoError(t, err)
		assert.Equal(t, someVersion, ver)

		res, err := e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		require.Len(t, res.Leases, 1)
		oldLeaseID := res.Leases[0].ID

		_, err = e.Server.LeaseRevoke(ctx, &etcdserverpb.LeaseRevokeRequest{ID: oldLeaseID})
		require.NoError(t, err)

		ver, err = st.Stash(ctx, someModule, someVersion)
		require.NoError(t, err)
		assert.Equal(t, someVersion, ver)

		res, err = e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)
		require.Len(t, res.Leases, 1)
		assert.NotEqual(t, oldLeaseID, res.Leases[0].ID)
	})

	t.Run("unlock failure orphans session", func(t *testing.T) {
		endpoint, e, closeServer := startClosableEtcdServer(t)
		const (
			someModule  = "example.com/mod"
			someVersion = "v1.0.0"
		)

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
		})
		require.NoError(t, err)

		var activeSess atomic.Pointer[concurrency.Session]
		var etcdStasher *etcd
		var ok bool
		etcdStasher, ok = w(stasherFunc(func(_ context.Context, _, ver string) (string, error) {
			etcdStasher.mu.Lock()
			activeSess.Store(etcdStasher.session)
			etcdStasher.mu.Unlock()
			closeServer()
			return ver, nil
		})).(*etcd)
		require.True(t, ok)
		etcdStasher.unlockTimeout = 200 * time.Millisecond

		ctx := context.WithValue(t.Context(), testContextKey{}, struct{}{})
		ver, err := etcdStasher.Stash(ctx, someModule, someVersion)
		require.NoError(t, err)
		assert.Equal(t, someVersion, ver)

		etcdStasher.mu.Lock()
		sess := etcdStasher.session
		etcdStasher.mu.Unlock()
		assert.Nil(t, sess)

		prev := activeSess.Load()
		require.NotNil(t, prev)
		select {
		case <-prev.Done():
		case <-time.After(2 * time.Second):
			t.Fatal("expected orphaned session Done channel to close")
		}
	})
}
