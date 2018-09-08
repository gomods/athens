package env

import (
	"github.com/gobuffalo/envy"
)

// TraceExporterURL returns where the trace is stored to
func TraceExporterURL() string {
	return envy.Get("TRACE_EXPORTER", "http://0.0.0.0:14268")
}
