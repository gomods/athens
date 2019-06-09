package addons

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gomods/athens/pkg/download"
	"github.com/gomods/athens/pkg/log"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/storage"
	"github.com/sirupsen/logrus"
)

// TestPoolLogic ensures that no
// more than given workers are working
// at one time.
func TestPoolLogic(t *testing.T) {
	m := &mockPool{}
	workers := 5
	dp := WithPool(workers)(m)
	ctx := context.Background()
	m.ch = make(chan struct{})
	for i := 0; i < 10; i++ {
		go dp.List(ctx, "")
	}
	<-m.ch
	if m.num != workers {
		t.Fatalf("expected %d workers but got %v", workers, m.num)
	}
}

type mockPool struct {
	download.Protocol
	num int
	mu  sync.Mutex
	ch  chan struct{}
}

func (m *mockPool) List(ctx context.Context, mod string) ([]string, error) {
	m.mu.Lock()
	m.num++
	if m.num == 5 {
		m.ch <- struct{}{}
	}
	m.mu.Unlock()

	time.Sleep(time.Minute)
	return nil, nil
}

// TestPoolWrapper ensures all upstream methods
// are successfully called.
func TestPoolWrapper(t *testing.T) {
	lggr := log.New("none", logrus.DebugLevel)
	m := &mockDP{}
	dp := WithPool(1)(m)
	ctx := context.Background()
	mod := "pkg"
	ver := "v0.1.0"
	m.inputMod = mod
	m.inputVer = ver
	m.list = []string{"v0.0.0", "v0.1.0"}
	m.catalog = []paths.AllPathParams{
		{Module: "pkg", Version: "v0.1.0"},
	}
	givenList, err := dp.List(ctx, mod)
	if err != m.err {
		t.Fatalf("expected dp.List err to be %v but got %v", m.err, err)
	}
	if !reflect.DeepEqual(m.list, givenList) {
		t.Fatalf("dp.List: expected %v and %v to be equal", m.list, givenList)
	}
	m.info = []byte("info response")
	givenInfo, err := dp.Info(ctx, mod, ver, lggr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m.info, givenInfo) {
		t.Fatalf("dp.Info: expected %s and %s to be equal", m.info, givenInfo)
	}
	m.err = fmt.Errorf("mod err")
	_, err = dp.GoMod(ctx, mod, ver, lggr)
	if m.err.Error() != err.Error() {
		t.Fatalf("dp.GoMod: expected err to be `%v` but got `%v`", m.err, err)
	}
	_, err = dp.Zip(ctx, mod, ver, lggr)
	if m.err.Error() != err.Error() {
		t.Fatalf("dp.Zip: expected err to be `%v` but got `%v`", m.err, err)
	}
}

type mockDP struct {
	err      error
	list     []string
	info     []byte
	latest   *storage.RevInfo
	gomod    []byte
	zip      io.ReadCloser
	inputMod string
	inputVer string
	catalog  []paths.AllPathParams
}

// List implements GET /{module}/@v/list
func (m *mockDP) List(ctx context.Context, mod string) ([]string, error) {
	if m.inputMod != mod {
		return nil, fmt.Errorf("expected mod input %v but got %v", m.inputMod, mod)
	}
	return m.list, m.err
}

// Info implements GET /{module}/@v/{version}.info
func (m *mockDP) Info(ctx context.Context, mod, ver string, lggr log.Entry) ([]byte, error) {
	if m.inputMod != mod {
		return nil, fmt.Errorf("expected mod input %v but got %v", m.inputMod, mod)
	}
	if m.inputVer != ver {
		return nil, fmt.Errorf("expected ver input %v but got %v", m.inputVer, ver)
	}
	return m.info, m.err
}

// Latest implements GET /{module}/@latest
func (m *mockDP) Latest(ctx context.Context, mod string) (*storage.RevInfo, error) {
	if m.inputMod != mod {
		return nil, fmt.Errorf("expected mod input %v but got %v", m.inputMod, mod)
	}
	return m.latest, m.err
}

// GoMod implements GET /{module}/@v/{version}.mod
func (m *mockDP) GoMod(ctx context.Context, mod, ver string, lggr log.Entry) ([]byte, error) {
	if m.inputMod != mod {
		return nil, fmt.Errorf("expected mod input %v but got %v", m.inputMod, mod)
	}
	if m.inputVer != ver {
		return nil, fmt.Errorf("expected ver input %v but got %v", m.inputVer, ver)
	}
	return m.gomod, m.err
}

// Zip implements GET /{module}/@v/{version}.zip
func (m *mockDP) Zip(ctx context.Context, mod, ver string, lggr log.Entry) (io.ReadCloser, error) {
	if m.inputMod != mod {
		return nil, fmt.Errorf("expected mod input %v but got %v", m.inputMod, mod)
	}
	if m.inputVer != ver {
		return nil, fmt.Errorf("expected ver input %v but got %v", m.inputVer, ver)
	}
	return m.zip, m.err
}

// Version is a helper method to get Info, GoMod, and Zip together.
func (m *mockDP) Version(ctx context.Context, mod, ver string) (*storage.Version, error) {
	panic("skipped")
}
