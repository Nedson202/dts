package middleware

import (
	"net/http"
	"time"

	"github.com/nedson202/dts-go/pkg/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		logger.Info().
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Int("status", ww.statusCode).
			Dur("duration", duration).
			Msg("HTTP request")
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
