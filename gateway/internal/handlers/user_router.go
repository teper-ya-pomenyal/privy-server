package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	mw "github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
)

func (h *UserHandler) NewRouter(m *mw.MiddleWares, allowedOrigins []string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{"Authorization", "Content-Type", "Range", "Content-Range"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
	}))
	r.Use(middleware.Logger)

	r.Use(middleware.Recoverer)

	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/refresh", h.Refresh)

	r.Group(func(r chi.Router) {
		r.Use(m.Auth.Handle)

		r.Post("/logout", h.Logout)
	})
	return r
}
