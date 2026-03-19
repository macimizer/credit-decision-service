package httpjson

import (
	"log/slog"
	"net/http"
	"time"
)

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			ww := &logStatusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			logger.InfoContext(r.Context(),
				"http request handled",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}

type logStatusWriter struct {
	http.ResponseWriter
	status int
}

func (w *logStatusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
