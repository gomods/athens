package events

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gomods/athens/pkg/requestid"
	"github.com/stretchr/testify/require"
	"github.com/technosophos/moniker"
)

var pingTests = []struct {
	name string
	err  error
}{
	{
		name: "ping",
	},
	{
		name: "ping_err",
		err:  fmt.Errorf("could not ping"),
	},
}

func TestClientServerPing(t *testing.T) {
	for _, tc := range pingTests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockHook{err: tc.err}
			srv := httptest.NewServer(NewServer(hook))
			t.Cleanup(srv.Close)
			client := NewClient(srv.URL, nil)
			err := client.Ping(context.Background())
			checkErr(t, tc.err != nil, err)
		})
	}
}

var stashedTests = []struct {
	name string
	mod  string
	ver  string
	err  error
}{
	{
		name: "happy path",
		mod:  "github.com/gomods/athens",
		ver:  "v0.10.0",
	},
	{
		name: "stashed error",
		mod:  "mod",
		ver:  "ver",
		err:  fmt.Errorf("server error"),
	},
}

func TestClientServerStashed(t *testing.T) {
	for _, tc := range stashedTests {
		t.Run(tc.name, func(t *testing.T) {
			hook := &mockHook{err: tc.err}
			srv := httptest.NewServer(NewServer(hook))
			t.Cleanup(srv.Close)
			client := NewClient(srv.URL, nil)
			err := client.Stashed(context.Background(), "github.com/gomods/athens", "v0.10.0")
			if checkErr(t, tc.err != nil, err) {
				return
			}
			if tc.mod != hook.mod {
				t.Fatalf("expected module to be %q but got %q", tc.mod, hook.mod)
			}
			if tc.ver != hook.ver {
				t.Fatalf("expected version to be %q but got %q", tc.ver, hook.ver)
			}
		})
	}
}

func TestRequestIDPropagation(t *testing.T) {
	hook := &mockHook{}
	srv := httptest.NewServer(NewServer(hook))
	t.Cleanup(srv.Close)
	client := NewClient(srv.URL, nil)
	reqID := moniker.New().Name()
	ctx := requestid.SetInContext(context.Background(), reqID)
	err := client.Ping(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if reqID != hook.reqid {
		t.Fatalf("expected request id to be %q but got %q", reqID, hook.reqid)
	}
}

type mockHook struct {
	mod, ver string
	reqid    string
	err      error
}

func (mh *mockHook) Ping(ctx context.Context) error {
	mh.reqid = requestid.FromContext(ctx)
	return mh.err
}

func (mh *mockHook) Stashed(ctx context.Context, mod, ver string) error {
	mh.mod, mh.ver = mod, ver
	return mh.err
}

func checkErr(t *testing.T, wantErr bool, err error) bool {
	if wantErr {
		if err == nil {
			t.Fatal("expected an error but got nil")
		}
		return true
	}
	require.NoError(t, err)
	return false
}
