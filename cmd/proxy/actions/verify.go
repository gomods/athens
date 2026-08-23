package actions

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/paths"
	"github.com/gomods/athens/pkg/verify"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// defaultSumDB is used when no SumDBs are configured.
const defaultSumDB = "https://sum.golang.org"

// ValidateVerifyFlags enforces that -purge is only meaningful together with
// -verify-storage. -purge is not a general purge; it only acts on the mismatch
// set produced by the verify pass.
func ValidateVerifyFlags(verifyStorage, purge bool) error {
	if purge && !verifyStorage {
		return fmt.Errorf("-purge requires -verify-storage")
	}
	return nil
}

// RunVerify builds the configured storage backend and sweeps it for zips that
// disagree with the checksum database, writing a report to out. When purge is
// true, mismatches are deleted. It does not start the server.
func RunVerify(conf *config.Config, purge bool, out io.Writer) error {
	const op errors.Op = "actions.RunVerify"

	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	store, err := GetStorage(conf.StorageType, conf.Storage, conf.TimeoutDuration(), client)
	if err != nil {
		return errors.E(op, err)
	}
	sweepStore, ok := store.(verify.Store)
	if !ok {
		return errors.E(op, fmt.Errorf("storage backend %q does not support cataloging; verify-storage is unavailable", conf.StorageType))
	}

	sumdbURL := defaultSumDB
	if len(conf.SumDBs) > 0 && conf.SumDBs[0] != "" {
		sumdbURL = conf.SumDBs[0]
	}
	oracle := verify.NewSumDBOracle(sumdbURL, client)

	patterns := conf.NoSumPatterns
	skip := func(mod string) bool {
		for _, p := range patterns {
			if paths.MatchesPattern(p, mod) {
				return true
			}
		}
		return false
	}

	if _, err := verify.Sweep(context.Background(), sweepStore, oracle, skip, purge, out); err != nil {
		return errors.E(op, err)
	}
	return nil
}
