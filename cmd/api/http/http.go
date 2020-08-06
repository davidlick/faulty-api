package http

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	RateLimiter RateLimiter
	server      *http.Server
}

func (s *Server) Run() error {
	s.server = &http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      s.BuildRoutes(),
	}

	fmt.Println("server listening on :8080")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
