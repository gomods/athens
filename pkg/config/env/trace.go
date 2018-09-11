package env

import "os"

// TraceExporterURL returns where the trace is stored to
func TraceExporterURL() string {
	return os.Getenv("TRACE_EXPORTER")
}
