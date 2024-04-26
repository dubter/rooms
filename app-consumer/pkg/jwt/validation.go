package jwt

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorBody struct {
	Message string `json:"message"`
}

func ProcessError(w http.ResponseWriter, msg string, code int) {
	body := ErrorBody{
		Message: msg,
	}
	buf, _ := json.Marshal(body)

	w.WriteHeader(code)
	_, _ = w.Write(buf)
}

func Validate(tokenManager TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				ProcessError(w, "empty auth header", http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(header, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				ProcessError(w, "invalid auth header", http.StatusUnauthorized)
				return
			}

			tokenString := headerParts[1]
			if len(tokenString) == 0 {
				ProcessError(w, "token is empty", http.StatusUnauthorized)
				return
			}

			user, err := tokenManager.Parse(tokenString)
			if err != nil {
				ProcessError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			r.Header.Set("user_id", user.UserID)
			r.Header.Set("nickname", user.Nickname)

			next.ServeHTTP(w, r)
		})
	}
}
