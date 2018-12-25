package middleware

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// RequestLogger logs request params to standard output
// it should only be used during dev.
func RequestLogger(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, 0}
		h.ServeHTTP(rw, r)
		fmt.Println(
			fmtRequest(
				r.Method,
				r.URL.Path,
				rw.statusCode,
			),
		)
	}
	return http.HandlerFunc(f)
}

func fmtRequest(method, path string, statusCode int) string {
	if statusCode == 0 {
		statusCode = 200
	}
	status := fmt.Sprint(statusCode)
	switch {
	case statusCode < http.StatusBadRequest:
		status = color.GreenString("%v", status)
	case statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError:
		status = color.HiYellowString("%v", status)
	default:
		status = color.HiRedString("%v", status)
	}
	return fmt.Sprintf(
		"%v %v %v [%v]",
		color.CyanString("handler:"),
		method,
		path,
		status,
	)
}
