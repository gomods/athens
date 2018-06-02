package actions

import (
	"github.com/gomods/athens/pkg/cdn"
	"github.com/gomods/athens/pkg/eventlog"
)

// mergeDB merges diff into the module database.
//
// TODO: this is racey if multiple processes are running mergeDB (they will be!) in a few ways:
//
// 1. CDN updates that race to change the /list endpoint
// 2. races between CDN updates and module metadata updates. For example:
//		- Delete operation deletes from the CDN
//		- Add operation adds to the CDN and saves to the module metadata DB
//		- Delete operation adds tombstone to module metadata k/v store
//
// Both could be fixed by putting each 'for' loop into a (global) critical section
func mergeDB(originURL string, diff dbDiff, getter cdn.Getter, saver cdn.Saver, deleter cdn.Deleter) error {
	for _, added := range diff.Added {
		if _, err := getter.GetForVersion(added.Module, added.Version); err != nil {
			// the module/version already exists, is deprecated, or is
			// tombstoned, so nothing to do
			break
		}

		// download code from the origin
		data, err := cdn.Download(originURL, added.Module, added.Version)
		if err != nil {
			return err
		}

		// save module data to the CDN
		if err := saver.Save(added.Module, data); err != nil {
			return err
		}

		// save module metadata to the key/value store
		if err := saver.SaveMetadata(added.Module, added.Version, eventlog.OpAdd); err != nil {
			return err
		}
	}
	for _, deprecated := range diff.Deprecated {
		fromDB, err := getter.GetForVersion(deprecated.Module, deprecated.Version)
		if err != nil {
			return err
		}
		if fromDB.Op == eventlog.OpDel {
			break // can't deprecate something that's already deleted
		}
		// delete from the CDN
		if err := deleter.Delete(deprecated.Module, deprecated.Version); err != nil {
			return err
		}

		// add the tombstone to module metadata
		if err := saver.SaveMetadata(deprecated.Module, deprecated.Version, eventlog.OpDep); err != nil {
			return err
		}
	}
	for _, deleted := range diff.Deleted {
		// delete in the CDN
		if err := deleter.Delete(deleted.Module, deleted.Version); err != nil {
			return err
		}
		// add tombstone to module metadata
		if err := saver.SaveMetadata(deleted.Module, deleted.Version, eventlog.OpDel); err != nil {
			return err
		}
	}

	return nil

}
