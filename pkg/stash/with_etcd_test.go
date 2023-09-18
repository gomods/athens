package stash

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
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

func startEtcdServer(t *testing.T) (endpoint string, e *etcdembed.Etcd) {
	c := etcdembed.NewConfig()
	c.Dir = t.TempDir()
	c.ListenPeerUrls = []url.URL{{Host: "localhost:0"}}
	c.ListenClientUrls = []url.URL{{Host: "localhost:0"}}

	e, err := etcdembed.StartEtcd(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { e.Close() })

	deadline, ok := t.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(time.Until(deadline)):
		t.Fatal("etcd server startup duration exceeded deadline")
	}
	return e.Clients[0].Addr().String(), e
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

		ctx := context.WithValue(context.Background(), testContextKey{}, struct{}{})

		w, err := WithEtcd([]string{endpoint}, &mockChecker{
			t:           t,
			e:           e,
			wantModule:  someModule,
			wantVersion: someVersion,
			exists:      true,
		})
		require.NoError(t, err)

		newVersion, err := w(&mockStasher{
			t:           t,
			wantModule:  someModule,
			wantVersion: someVersion,
		}).Stash(ctx, someModule, someVersion)
		require.NoError(t, err)

		assert.Equal(t, someVersion, newVersion)

		res, err := e.Server.LeaseLeases(ctx, &etcdserverpb.LeaseLeasesRequest{})
		require.NoError(t, err)

		assert.Empty(t, res.Leases)
	})

	t.Run("module does not exist", func(t *testing.T) {
		endpoint, e := startEtcdServer(t)

		const (
			someModule  = "some module"
			someVersion = "some version"

			someNewVersion = "some new version"
		)

		ctx := context.WithValue(context.Background(), testContextKey{}, struct{}{})

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

		ctx := context.WithValue(context.Background(), testContextKey{}, struct{}{})

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

		ctx := context.WithValue(context.Background(), testContextKey{}, struct{}{})

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
}
