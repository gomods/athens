package env

import (
	"runtime"
	"strconv"

	"github.com/gobuffalo/envy"
)

// OlympusGlobalEndpointWithDefault returns Olympus global endpoint defined by OLYMPUS_GLOBAL_ENDPOINT.
func OlympusGlobalEndpointWithDefault(value string) string {
	return envy.Get("OLYMPUS_GLOBAL_ENDPOINT", value)
}

// AthensMaxConcurrency retrieves maximal level of concurrency based on ATHENS_MAX_CONCURRENCY.
// Defaults to number of cores if env is not set.
func AthensMaxConcurrency() int {
	defaultMaxConcurrency := runtime.NumCPU()
	maxConcurrencyEnv, err := envy.MustGet("ATHENS_MAX_CONCURRENCY")
	if err != nil {
		return defaultMaxConcurrency
	}

	mc, err := strconv.Atoi(maxConcurrencyEnv)
	if err != nil {
		return defaultMaxConcurrency
	}

	return mc
}
