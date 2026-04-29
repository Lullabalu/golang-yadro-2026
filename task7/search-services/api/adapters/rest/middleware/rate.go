package middleware

import (
	"golang.org/x/time/rate"
	"net/http"
)

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), 1)
	return func(w http.ResponseWriter, r *http.Request) {
		if err := limiter.Wait(r.Context()); err != nil {
			http.Error(w, "Сервис временно недоступен", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	}
}
