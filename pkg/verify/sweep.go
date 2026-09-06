package verify

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gomods/athens/pkg/storage"
)

// pageSize is how many catalog entries to fetch per page.
const pageSize = 1000

// Store is the subset of a storage backend the sweep needs.
type Store interface {
	storage.Cataloger
	storage.Getter
	storage.Deleter
}

// Report tallies the outcome of a sweep.
type Report struct {
	Matched        int // stored hash == canonical
	Mismatched     int // stored hash != canonical
	Purged         int // mismatches deleted (purge mode)
	SkippedPrivate int // matched a NoSumPattern, never looked up
	SkippedUnknown int // absent from the checksum DB (ErrNotInSumDB)
	Unverified     int // transient/storage/oracle error; kept
	DeleteErrors   int // Delete failed in purge mode
}

// Sweep verifies every catalogued module version against the oracle. skip
// reports whether a module is private and MUST be consulted before any oracle
// lookup (to avoid leaking private module paths). When purge is true,
// mismatches are deleted; otherwise they are only reported.
func Sweep(ctx context.Context, s Store, o Oracle, skip func(mod string) bool, purge bool, out io.Writer) (Report, error) {
	var rep Report
	token := ""
	for {
		page, next, err := s.Catalog(ctx, token, pageSize)
		if err != nil {
			return rep, fmt.Errorf("verify: catalog: %w", err)
		}
		for _, mv := range page {
			if skip(mv.Module) {
				rep.SkippedPrivate++
				continue
			}
			mismatch, err := Poisoned(ctx, s, o, mv.Module, mv.Version)
			switch {
			case errors.Is(err, ErrNotInSumDB):
				rep.SkippedUnknown++
				continue
			case err != nil:
				rep.Unverified++
				fmt.Fprintf(out, "unverified\t%s@%s\t%v\n", mv.Module, mv.Version, err)
				continue
			case !mismatch:
				rep.Matched++
				continue
			}
			rep.Mismatched++
			fmt.Fprintf(out, "MISMATCH\t%s@%s\n", mv.Module, mv.Version)
			if purge {
				if err := s.Delete(ctx, mv.Module, mv.Version); err != nil {
					rep.DeleteErrors++
					fmt.Fprintf(out, "delete-failed\t%s@%s\t%v\n", mv.Module, mv.Version, err)
					continue
				}
				rep.Purged++
			}
		}
		if next == "" {
			break
		}
		token = next
	}
	fmt.Fprintf(out, "\nmatched=%d mismatched=%d purged=%d skipped_private=%d skipped_unknown=%d unverified=%d delete_errors=%d\n",
		rep.Matched, rep.Mismatched, rep.Purged, rep.SkippedPrivate, rep.SkippedUnknown, rep.Unverified, rep.DeleteErrors)
	return rep, nil
}
