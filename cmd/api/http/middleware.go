package http

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
)

func (s *Server) rateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := s.RateLimiter.Acquire()
		if err != nil {
			panic(err)
		}
		ctx := context.WithValue(r.Context(), "rate-limit-token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) faulty(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rlExceedPerc := s.RateLimiter.LimitExceededPerc()
		// Faults will scale to the percentage of requests that are above the limit.
		// When 2x volume is exceeded all requests should fail.
		if rlExceedPerc > 0.0 {
			check := rand.Float32()
			if rlExceedPerc > float32(check) {
				s.respondError(w, http.StatusInternalServerError, errors.New("application error"))
				// Transaction failed so release the token.
				s.RateLimiter.Release(r.Context().Value("rate-limit-token").(*Token))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
