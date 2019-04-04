package mem

import (
	"fmt"
	"sync"
	"time"

	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/spf13/afero"
)

var (
	l          sync.Mutex
	memStorage *stg
)

type stg struct {
	listCacheMut sync.RWMutex
	listCache    map[string]time.Time
	ttl          time.Duration
	storage.Backend
}

// NewStorage creates new in-memory storage using the afero.NewMemMapFs() in memory file system
func NewStorage() (storage.Backend, error) {
	const op errors.Op = "mem.NewStorage"
	l.Lock()
	defer l.Unlock()

	if memStorage != nil {
		return memStorage, nil
	}

	memFs := afero.NewMemMapFs()
	tmpDir, err := afero.TempDir(memFs, "", "")
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not create temp dir for 'In Memory' storage (%s)", err))
	}

	backendImpl, err := fs.NewStorage(tmpDir, memFs)
	if err != nil {
		return nil, errors.E(op, fmt.Errorf("could not create storage from memory fs (%s)", err))
	}

	// TODO: pass this through from config
	const cacheTTL = 100 * time.Millisecond
	return &stg{
		Backend:   backendImpl,
		listCache: map[string]time.Time{},
		ttl:       cacheTTL,
	}, nil
}
