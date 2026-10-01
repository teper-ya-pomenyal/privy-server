package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/teper-ya-pomenyal/privy_stream/jwtmanager"
)

type contextKey string

const UserIDKey contextKey = "userID"

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
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, BDKey, claims.BirthDate)
		ctx = context.WithValue(ctx, roleKey, claims.Role)

		newReq := r.WithContext(ctx)
		next.ServeHTTP(w, newReq)

	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok && role != ""
}

func (m *authMiddleware) StreamAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.URL.Query().Get("access_token")
		if tokenString == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := m.verifier.VerifyStreamToken(tokenString)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Scope != "stream" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), BDKey, claims.BirthDate))
		next.ServeHTTP(w, r)
	})
}
