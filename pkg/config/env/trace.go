package env

import "os"

// TraceExporterURL returns where the trace is stored to
func TraceExporterURL() string {
	url := os.Getenv("TRACE_EXPORTER")
	if url == "" {
		return "http://0.0.0.0:14268"
	}
	return url
}
