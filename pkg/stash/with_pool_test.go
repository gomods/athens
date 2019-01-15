package stash

import (
	"context"
	"fmt"
	"testing"
)

func TestPoolWrapper(t *testing.T) {
	m := &mockPoolStasher{inputMod: "mod", inputVer: "ver", err: fmt.Errorf("wrapped err")}
	s := WithPool(2)(m)
	_, err := s.Stash(context.Background(), m.inputMod, m.inputVer)
	if err.Error() != m.err.Error() {
		t.Fatalf("expected err to be `%v` but got `%v`", m.err, err)
	}
}

type mockPoolStasher struct {
	inputMod string
	inputVer string
	err      error
}

func (m *mockPoolStasher) Stash(ctx context.Context, mod, ver string) (string, error) {
	if m.inputMod != mod {
		return "", fmt.Errorf("expected input mod %v but got %v", m.inputMod, mod)
	}
	if m.inputVer != ver {
		return "", fmt.Errorf("expected input ver %v but got %v", m.inputVer, ver)
	}
	return "", m.err
}
