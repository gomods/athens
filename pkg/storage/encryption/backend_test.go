package encryption

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/gomods/athens/pkg/storage"
)

func TestMustWrapPanicKeyTooSmall(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic from key being too small")
		}
	}()
	MustWrap(&mockBackend{}, "")
}

func TestMustWrapPanicKeyTooBig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic from key being too big")
		}
	}()
	MustWrap(&mockBackend{}, "123456789012345678901234567890123")
}

func TestMustWrapNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("expected no panic from key with correct length")
		}
	}()
	MustWrap(&mockBackend{}, "12345678901234567890123456789012")
}

func TestWrap(t *testing.T) {

	tests := []struct {
		desc        string
		key         string
		expectedErr bool
	}{
		{
			desc:        "key empty",
			key:         "",
			expectedErr: true,
		},
		{
			desc:        "key too small",
			key:         "1234567890123456789012345678901",
			expectedErr: true,
		},
		{
			desc:        "key too big",
			key:         "123456789012345678901234567890123",
			expectedErr: true,
		},
		{
			desc:        "key with correct length",
			key:         "12345678901234567890123456789012",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {

			_, err := Wrap(&mockBackend{}, tt.key)
			if tt.expectedErr && err == nil {
				t.Error("expected error, but got nil")
			} else if !tt.expectedErr && err != nil {
				t.Errorf("expected no error, but got %v", err)
			}

		})
	}

}

func TestPassThroughMethods(t *testing.T) {

	key := "12345678901234567890123456789012"
	ctx := context.TODO()
	mod := "mod"
	ver := "ver"
	vers := []string{ver}
	mockErr := errors.New("mock error")

	if _, err := MustWrap(&mockBackend{
		list: func(_ context.Context, _ string) ([]string, error) {
			return nil, mockErr
		},
	}, key).List(ctx, mod); err == nil {
		t.Errorf("expected List method to pass through error, but got nil")
	}

	versions, err := MustWrap(&mockBackend{
		list: func(_ context.Context, _ string) ([]string, error) {
			return vers, nil
		},
	}, key).List(ctx, mod)

	if err != nil {
		t.Errorf("expected List method to pass through versions, but got error: %v", err)
	}
	if !reflect.DeepEqual(versions, vers) {
		t.Errorf("expected List method to pass through versions, got %v wanted %v", versions, vers)
	}

	if _, err := MustWrap(&mockBackend{
		exists: func(_ context.Context, _ string, _ string) (bool, error) {
			return false, mockErr
		},
	}, key).Exists(ctx, mod, ver); err == nil {
		t.Errorf("expected Exists method to pass through error, but got nil")
	}

	exists, err := MustWrap(&mockBackend{
		exists: func(_ context.Context, _ string, _ string) (bool, error) {
			return true, nil
		},
	}, key).Exists(ctx, mod, ver)

	if err != nil {
		t.Errorf("expected Exists method to pass through bool, but got error: %v", err)
	}
	if !reflect.DeepEqual(exists, true) {
		t.Errorf("expected Exists method to pass through bool, got %t wanted %t", exists, true)
	}

	if err := MustWrap(&mockBackend{
		del: func(_ context.Context, _ string, _ string) error {
			return mockErr
		},
	}, key).Delete(ctx, mod, ver); err == nil {
		t.Errorf("expected Delete method to pass through error, but got nil")
	}

	if err := MustWrap(&mockBackend{
		del: func(_ context.Context, _ string, _ string) error {
			return nil
		},
	}, key).Delete(ctx, mod, ver); err != nil {
		t.Errorf("expected Delete method to pass through nil, but got error: %v", err)
	}

}

// func TestMustWrapBackendDecryptingMethods(t *testing.T) {

// 	key := "change this password to a secret"
// 	ctx := context.TODO()
// 	mod := "mod"
// 	ver := "ver"
// 	i := []byte("c3aaa29f002ca75870806e44086700f62ce4d43e902b3888e23ceff797a7a471")
// 	mockErr := errors.New("mock error")

// 	if _, err := MustWrap(&mockBackend{
// 		info: func(_ context.Context, _ string, _ string) ([]byte, error) {
// 			return nil, mockErr
// 		},
// 	}, key).Info(ctx, mod, ver); err == nil {
// 		t.Errorf("expected Info method to pass through error, but got nil")
// 	}

// 	info, err := MustWrap(&mockBackend{
// 		info: func(_ context.Context, _ string, _ string) ([]byte, error) {
// 			return i, nil
// 		},
// 	}, key).Info(ctx, mod, ver)

// 	if err != nil {
// 		t.Errorf("expected Info method to pass through decrypted info, but got error: %v", err)
// 	}
// 	if string(info) != string(i) {
// 		t.Errorf("expected Info method to decrypt info, got %s wanted %s", string(info), string(i))
// 	}

// }

func TestSave(t *testing.T) {

	key := "12345678901234567890123456789012"
	mockRandReader := bytes.NewReader([]byte(key + key + key + key + key + key))
	ctx := context.Background()
	mod := "mod"
	ver := "ver"
	zip := bytes.NewBuffer([]byte(""))
	info := "info"
	var b storage.Backend

	expectedCipherMod := "3334353637383930313233347baf4c2c7dda36fa30feccc91141ccb1950621"
	expectedCipherInfo := "353637383930313231323334d70472bed246db8d48d75f5a51e762d518069fe2"

	b = MustWrap(&mockBackend{
		save: func(_ context.Context, _ string, _ string, _ []byte, _ io.Reader, _ []byte) error {
			return errors.New("mock error")
		},
	}, key)

	if err := b.Save(ctx, mod, ver, []byte(mod), zip, []byte(info)); err == nil {
		t.Errorf("expected Save method err; got nil")
	}

	b = MustWrap(&mockBackend{
		save: func(_ context.Context, _ string, _ string, mod []byte, _ io.Reader, info []byte) error {

			if fmt.Sprintf("%x", mod) != expectedCipherMod {
				t.Errorf("expected Save method to pass through encrypted mod, got %x, wanted %s", mod, expectedCipherMod)
			}
			if fmt.Sprintf("%x", info) != expectedCipherInfo {
				t.Errorf("expected Save method to pass through encrypted info, got %x, wanted %s", info, expectedCipherInfo)
			}

			return nil
		},
	}, key)
	eb := b.(*backend)
	eb.randReader = mockRandReader

	if err := eb.Save(ctx, mod, ver, []byte(mod), zip, []byte(info)); err != nil {
		t.Errorf("expected Save method not err; got %v", err)
	}
}

func TestInfo(t *testing.T) {

	key := "12345678901234567890123456789012"
	ctx := context.Background()
	mod := "mod"
	ver := "ver"
	var b storage.Backend

	cipherInfo, _ := hex.DecodeString("353637383930313231323334d70472bed246db8d48d75f5a51e762d518069fe2")
	expectedPlaintextInfo := "info"

	b = MustWrap(&mockBackend{
		info: func(_ context.Context, _ string, _ string) ([]byte, error) {
			return nil, errors.New("mock error")
		},
	}, key)

	if _, err := b.Info(ctx, mod, ver); err == nil {
		t.Errorf("expected Info method err; got nil")
	}

	b = MustWrap(&mockBackend{
		info: func(_ context.Context, _ string, _ string) ([]byte, error) {
			return []byte(cipherInfo), nil
		},
	}, key)

	actualInfo, err := b.Info(ctx, mod, ver)
	if err != nil {
		t.Errorf("expected Info method not err; got %v", err)
	}

	if string(actualInfo) != expectedPlaintextInfo {
		t.Errorf("expected Info method to pass through decrypted info, got %x, wanted %s", string(actualInfo), expectedPlaintextInfo)
	}
}

type mockBackend struct {
	list   func(context.Context, string) ([]string, error)
	exists func(context.Context, string, string) (bool, error)
	info   func(context.Context, string, string) ([]byte, error)
	goMod  func(context.Context, string, string) ([]byte, error)
	zip    func(context.Context, string, string) (io.ReadCloser, error)
	save   func(context.Context, string, string, []byte, io.Reader, []byte) error
	del    func(context.Context, string, string) error
}

func (b *mockBackend) List(ctx context.Context, module string) ([]string, error) {
	return b.list(ctx, module)
}

func (b *mockBackend) Exists(ctx context.Context, module, version string) (bool, error) {
	return b.exists(ctx, module, version)
}

func (b *mockBackend) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	return b.info(ctx, module, vsn)
}

func (b *mockBackend) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	return b.goMod(ctx, module, vsn)
}

func (b *mockBackend) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	return b.zip(ctx, module, vsn)
}

func (b *mockBackend) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	return b.save(ctx, module, version, mod, zip, info)
}

func (b *mockBackend) Delete(ctx context.Context, module, vsn string) error {
	return b.del(ctx, module, vsn)
}
