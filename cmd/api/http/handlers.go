package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

func (s *Server) setLimit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Limit int `json:"limit"`
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)
		return
	}

	err = json.Unmarshal(bodyBytes, &input)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)
	}

	var rl RateLimiter
	rl, err = NewMaxConcurrencyRateLimiter(input.Limit)
	if err != nil {
		panic(err)
	}
	fmt.Printf("setting new rate limiter: %+v\n", rl)
	s.RateLimiter = rl
}

func (s *Server) getData(w http.ResponseWriter, r *http.Request) {
	data := []struct {
		Item1 string
		Item2 string
		Item3 string
	}{
		{
			Item1: "books",
			Item2: "hotdogs",
			Item3: "trinkets",
		},
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err)
		return
	}

	// Sleep for a random amount of time up to 500 milliseconds.
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	w.Write(bytes)

	t := r.Context().Value("rate-limit-token").(*Token)
	s.RateLimiter.Release(t)
	return
}

func (s *Server) respondError(w http.ResponseWriter, statusCode int, err error) {
	http.Error(w, err.Error(), statusCode)
	return
}
