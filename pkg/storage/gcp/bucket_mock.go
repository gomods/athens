package gcp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

type bucketMock struct {
	db   map[string][]byte
	lock sync.RWMutex
}

func newBucketMock() Bucket {
	b := &bucketMock{}
	b.db = make(map[string][]byte)
	return b
}

type bucketReader struct {
	io.Reader
	*bucketMock
}

type bucketWriter struct {
	*bucketMock
	path string
}

func (r *bucketReader) Close() error {
	r.bucketMock.lock.RUnlock()
	return nil
}

func (r *bucketWriter) Close() error {
	r.bucketMock.lock.Unlock()
	return nil
}

func (r *bucketWriter) Write(p []byte) (int, error) {
	_, ok := r.bucketMock.db[r.path]
	if !ok {
		r.bucketMock.db[r.path] = make([]byte, 0)
	}

	r.bucketMock.db[r.path] = append(r.bucketMock.db[r.path], p...)

	return len(p), nil
}

func (m *bucketMock) Delete(ctx context.Context, path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.db, path)
	return nil
}

func (m *bucketMock) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	data, ok := m.db[path]
	if !ok {
		return nil, fmt.Errorf("path %s not found", path)
	}

	m.lock.RLock()
	r := bytes.NewReader(data)
	return &bucketReader{r, m}, nil
}

func (m *bucketMock) Write(ctx context.Context, path string) io.WriteCloser {
	m.lock.Lock()
	return &bucketWriter{m, path}
}

func (m *bucketMock) List(ctx context.Context, prefix string) ([]string, error) {
	res := make([]string, 0)

	m.lock.RLock()
	defer m.lock.RUnlock()
	for k := range m.db {
		if strings.HasPrefix(k, prefix) {
			res = append(res, k)
		}
	}
	return res, nil
}

func (m *bucketMock) Exists(ctx context.Context, path string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, found := m.db[path]
	return found
}
