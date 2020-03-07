package encryption

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/gomods/athens/pkg/storage"
)

// MustWrap will wrap an athen's storage.Backend with an encrypt/decrypt layer and panic if there is an error.
func MustWrap(store storage.Backend, key string) storage.Backend {
	b, err := Wrap(store, key)
	if err != nil {
		panic(err)
	}
	return b
}

// Wrap will wrap an athen's storage.Backend with an encrypt/decrypt layer.
func Wrap(store storage.Backend, key string) (storage.Backend, error) {

	if len(key) != 32 {
		return nil, fmt.Errorf("length of encryption key must be 32; found key of length: %d", len(key))
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &backend{
		Store: store,
		AEAD:  aead,
	}, nil
}

// backend defines a backend that is wrapped with an encrypt/decrypt layer.
type backend struct {
	Store storage.Backend
	AEAD  cipher.AEAD

	initOnce   sync.Once
	randReader io.Reader
}

func (b *backend) init() {
	if b.randReader == nil {
		b.randReader = rand.Reader
	}
}

func (b *backend) List(ctx context.Context, module string) ([]string, error) {
	b.initOnce.Do(b.init)
	return b.Store.List(ctx, module)
}

func (b *backend) Exists(ctx context.Context, module, version string) (bool, error) {
	b.initOnce.Do(b.init)
	return b.Store.Exists(ctx, module, version)
}

func (b *backend) Info(ctx context.Context, module, vsn string) ([]byte, error) {
	b.initOnce.Do(b.init)
	return b.open(b.Store.Info(ctx, module, vsn))
}

func (b *backend) GoMod(ctx context.Context, module, vsn string) ([]byte, error) {
	b.initOnce.Do(b.init)
	return b.open(b.Store.GoMod(ctx, module, vsn))
}

func (b *backend) Zip(ctx context.Context, module, vsn string) (io.ReadCloser, error) {
	b.initOnce.Do(b.init)
	return b.openReadCloserer(b.Store.Zip(ctx, module, vsn))
}

func (b *backend) Save(ctx context.Context, module, version string, mod []byte, zip io.Reader, info []byte) error {
	b.initOnce.Do(b.init)
	zipReader, err := b.sealReader(zip)
	if err != nil {
		return err
	}
	sealedMod, err := b.seal(mod)
	if err != nil {
		return err
	}
	sealedInfo, err := b.seal(info)
	if err != nil {
		return err
	}
	return b.Store.Save(ctx, module, version, sealedMod, zipReader, sealedInfo)
}

func (b *backend) Delete(ctx context.Context, module, vsn string) error {
	b.initOnce.Do(b.init)
	return b.Store.Delete(ctx, module, vsn)
}

func (b *backend) seal(data []byte) ([]byte, error) {
	nonce := make([]byte, b.AEAD.NonceSize())

	_, err := io.ReadFull(b.randReader, nonce)
	if err != nil {
		return nil, err
	}
	encryptedData := b.AEAD.Seal(nil, nonce, data, nil)

	return bytes.Join([][]byte{nonce, encryptedData}, nil), nil
}

func (b *backend) sealReader(w io.Reader) (io.Reader, error) {

	plaintextData, err := ioutil.ReadAll(w)
	if err != nil {
		return nil, err
	}
	d, err := b.seal(plaintextData)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(d), nil
}

func (b *backend) open(cipherdata []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	nonceSize := b.AEAD.NonceSize()
	if len(cipherdata) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, cipherdata := cipherdata[:nonceSize], cipherdata[nonceSize:]
	return b.AEAD.Open(nil, nonce, cipherdata, nil)
}

func (b *backend) openReadCloserer(rc io.ReadCloser, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}
	plaintextdata, err := b.open(ioutil.ReadAll(rc))
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.NopCloser(bytes.NewBuffer(plaintextdata)), nil
}
