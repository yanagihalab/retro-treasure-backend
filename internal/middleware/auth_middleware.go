package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"retro-treasure-backend/internal/repository"
)

type contextKey string

const UserIDKey contextKey = "userID"

func RequireAuth(repo *repository.MemoryRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		userID, err := repo.UserIDByToken(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(UserIDKey)
	id, ok := v.(int64)
	return id, ok
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
