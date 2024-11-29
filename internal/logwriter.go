package internal

import (
	"log/slog"
	"net/http"
)

type logWriter struct {
	http.ResponseWriter
	code int
}

func (lw *logWriter) WriteHeader(code int) {
	lw.code = code
	lw.ResponseWriter.WriteHeader(code)
}

// Unwrap supports http.ResponseController.
func (lw *logWriter) Unwrap() http.ResponseWriter { return lw.ResponseWriter }

func LogWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &logWriter{w, http.StatusOK}
		next.ServeHTTP(lw, r)
		slog.Debug(r.Method+"\t"+r.URL.Path, "code", lw.code)
	})
}
