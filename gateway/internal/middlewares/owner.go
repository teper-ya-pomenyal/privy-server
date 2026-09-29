package middlewares

import (
	"net/http"
)

// RequireOwner: административные операции (создание/удаление каталога и
// будущие admin-маршруты) разрешены только роли owner из claims access-токена.
// Роль кладёт authMiddleware, поэтому RequireOwner всегда идёт после него.
type ownerMiddleware struct{}

func (m *ownerMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := RoleFromContext(r.Context())
		if !ok || role != "owner" {
			http.Error(w, "owner role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
