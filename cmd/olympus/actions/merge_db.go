package actions

import (
	"context"
	"log"
	"time"

	"github.com/gomods/athens/pkg/cdn"
	"github.com/gomods/athens/pkg/eventlog"
	"github.com/gomods/athens/pkg/storage"
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
func mergeDB(ctx context.Context, originURL string, diff dbDiff, eLog eventlog.Eventlog, storage storage.Backend) error {
	for _, added := range diff.Added {
		err := func(ae eventlog.Event) error {
			if _, err := eLog.ReadSingle(ae.Module, ae.Version); err != nil {
				// the module/version already exists, is deprecated, or is
				// tombstoned, so nothing to do
				return err
			}

			// download code from the origin
			data, err := cdn.Download(originURL, ae.Module, ae.Version)
			if err != nil {
				log.Printf("error downloading new module %s/%s from %s (%s)", ae.Module, ae.Version, originURL, err)
				return err
			}
			defer data.Zip.Close()
			// save module data to the CDN
			if err := storage.Save(ctx, ae.Module, ae.Version, data.Mod, data.Zip, data.Info); err != nil {
				log.Printf("error saving new module %s/%s to CDN (%s)", ae.Module, ae.Version, err)
				return err
			}

			// save module metadata to the key/value store
			if _, err := eLog.Append(eventlog.Event{Module: ae.Module, Version: ae.Version, Time: time.Now(), Op: eventlog.OpAdd}); err != nil {
				log.Printf("error saving metadata for new module %s/%s (%s)", ae.Module, ae.Version, err)
				return err
			}
			return nil
		}(added)
		if err != nil {
			continue
		}
	}
	for _, deprecated := range diff.Deprecated {
		fromDB, err := eLog.ReadSingle(deprecated.Module, deprecated.Version)
		if err != nil {
			log.Printf("error getting deprecated module %s/%s (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}
		if fromDB.Op == eventlog.OpDel {
			continue // can't deprecate something that's already deleted
		}
		// delete from the CDN
		if err := storage.Delete(deprecated.Module, deprecated.Version); err != nil {
			log.Printf("error deleting deprecated module %s/%s from CDN (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}

		// add the tombstone to module metadata
		if _, err := eLog.Append(eventlog.Event{Module: deprecated.Module, Version: deprecated.Version, Time: time.Now(), Op: eventlog.OpDel}); err != nil {
			log.Printf("error saving metadata for deprecated module %s/%s from CDN (%s)", deprecated.Module, deprecated.Version, err)
			continue
		}
	}
	for _, deleted := range diff.Deleted {
		// delete in the CDN
		if err := storage.Delete(deleted.Module, deleted.Version); err != nil {
			log.Printf("error deleting deleted module %s/%s from CDN (%s)", deleted.Module, deleted.Version, err)
			continue
		}
		// add tombstone to module metadata
		if _, err := eLog.Append(eventlog.Event{Module: deleted.Module, Version: deleted.Version, Time: time.Now(), Op: eventlog.OpDel}); err != nil {
			log.Printf("error inserting tombstone for deleted module %s/%s (%s)", deleted.Module, deleted.Version, err)
			return err
		}
	}

	return nil

}
