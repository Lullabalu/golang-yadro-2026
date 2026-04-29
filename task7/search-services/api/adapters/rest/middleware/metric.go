package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{
			ResponseWriter: w,
			code:           http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf("method=%s path=%s status=%d duration=%s",
			r.Method,
			r.URL.Path,
			rw.code,
			duration,
		)
	})
}
