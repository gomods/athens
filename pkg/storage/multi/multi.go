package multi

import (
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/storage"
	multierror "github.com/hashicorp/go-multierror"
)

// Storage is a storage.Backend implementations
// combining storage.Backend interfaces
// serving first to respond with data for reading
// waits all while writing
type Storage struct {
	storages []storage.Backend
}

// NewStorage creates an instance of MultiStorage
func NewStorage(storages []storage.Backend) (*Storage, error) {
	return &Storage{
		storages: storages,
	}, nil
}

func (s *Storage) composeError(module, version string, op errors.Op, errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	var err error
	err = multierror.Append(err, errs...)
	// if there was at least 1 failure
	if len(errs) != len(s.storages) {
		return err
	}

	// if all nil, return nil
	var notNilFound bool
	for _, e := range errs {
		if e != nil {
			notNilFound = true
			break
		}
	}
	if !notNilFound {
		return nil
	}

	// if at least one is not not found, return multi errors
	for _, e := range errs {
		if errors.KindNotFound != errors.Kind(e) {
			return err
		}
	}

	// all are not found, return flattened
	if version == "" {
		return errors.E(op, errors.M(module), errors.KindNotFound)
	}

	return errors.E(op, errors.M(module), errors.V(version), errors.KindNotFound)
}
