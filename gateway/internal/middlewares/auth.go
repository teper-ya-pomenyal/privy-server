package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/teper-ya-pomenyal/privy_stream/jwtmanager"
)

type contextKey string

const userIDKey contextKey = "userID"

type contextBirthDate string

const BDKey contextBirthDate = "birthDate"

type contextRole string

const roleKey contextRole = "role"

type authMiddleware struct {
	verifier *jwtmanager.Verifier
}

func (m *authMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(b, "Bearer ")
		if !ok {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		claims, err := m.verifier.VerifyAccessToken(token)
		if err != nil {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
		ctx = context.WithValue(ctx, BDKey, claims.BirthDate)
		ctx = context.WithValue(ctx, roleKey, claims.Role)

		newReq := r.WithContext(ctx)
		next.ServeHTTP(w, newReq)

	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok && role != ""
}
