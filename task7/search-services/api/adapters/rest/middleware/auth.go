package middleware

import (
	"errors"
	"net/http"
	"strings"

	"yadro.com/course/api/core"
)

type TokenVerifier interface {
	Verify(token string) error
}

func Auth(next http.HandlerFunc, verifier TokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Token ")
		err := verifier.Verify(tokenString)
		if err != nil {
			if errors.Is(err, core.ErrForbidden) {
				http.Error(w, "Нет доступа", http.StatusForbidden)
			} else {
				http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
			}
			return
		}

		next(w, r)
	}
}
