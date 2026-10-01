package handlers

import (
	"log"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
	mw "github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
)

func (h *StreamHandlers) NewStreamingRouter(address string, m *mw.MiddleWares) *chi.Mux {
	target, err := url.Parse("http://" + address)
	if err != nil {
		log.Fatalln(err)
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)

			if bd, ok := pr.In.Context().Value(middlewares.BDKey).(time.Time); ok {
				pr.Out.Header.Set("X-Birth-Date", bd.Format(time.DateOnly))
			}
		},
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Group(func(r chi.Router) {
		r.Use(m.Auth.Handle)
		r.Post("/stream/token", h.CreateStreamToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(m.Auth.StreamAuth)
		r.Get("/stream/{id}", proxy.ServeHTTP)
	})
	return r
}
