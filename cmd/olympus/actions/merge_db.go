package actions

import (
	"log"

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
			continue
		}

		// download code from the origin
		data, err := cdn.Download(originURL, added.Module, added.Version)
		if err != nil {
			log.Printf("error downloading new module %s/%s from %s (%s)", added.Module, added.Version, originURL, err)
			continue
		}

		// save module data to the CDN
		if err := saver.Save(added.Module, data); err != nil {
			log.Printf("error saving new module %s/%s to CDN (%s)", added.Module, added.Version, err)
			continue
		}

		// save module metadata to the key/value store
		if err := saver.SaveMetadata(added.Module, added.Version, eventlog.OpAdd); err != nil {
			log.Printf("error saving metadata for new module %s/%s (%s)", added.Module, added.Version, err)
			continue
		}
	}
	for _, deprecated := range diff.Deprecated {
		fromDB, err := getter.GetForVersion(deprecated.Module, deprecated.Version)
		if err != nil {
			log.Printf("error getting deprecated module %s/%s (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}
		if fromDB.Op == eventlog.OpDel {
			continue // can't deprecate something that's already deleted
		}
		// delete from the CDN
		if err := deleter.Delete(deprecated.Module, deprecated.Version); err != nil {
			log.Printf("error deleting deprecated module %s/%s from CDN (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}

		// add the tombstone to module metadata
		if err := saver.SaveMetadata(deprecated.Module, deprecated.Version, eventlog.OpDep); err != nil {
			log.Printf("error saving metadata for deprecated module %s/%s from CDN (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}
	}
	for _, deleted := range diff.Deleted {
		// delete in the CDN
		if err := deleter.Delete(deleted.Module, deleted.Version); err != nil {
			log.Printf("error deleting deleted module %s/%s from CDN (%s)", deleted.Module, deleted.Version, err)
			continue
		}
		// add tombstone to module metadata
		if err := saver.SaveMetadata(deleted.Module, deleted.Version, eventlog.OpDel); err != nil {
			log.Printf("error inserting tombstone for deleted module %s/%s (%s)", deleted.Module, deleted.Version, err)
			return err
		}
	}

	return nil

}
