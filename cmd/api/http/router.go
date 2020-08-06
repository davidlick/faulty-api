package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (s *Server) BuildRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/limit", s.setLimit)

	r.Route("/data", func(r chi.Router) {
		r.Use(s.rateLimiter)
		r.Use(s.faulty)

		r.Get("/", s.getData)
	})

	return r
}
