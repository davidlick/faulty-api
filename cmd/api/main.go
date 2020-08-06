package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RedVentures/agentx-loadtest-lunchandlearn/cmd/api/http"
)

func main() {
	rl, err := http.NewMaxConcurrencyRateLimiter(10)
	if err != nil {
		log.Fatal(err)
	}
	s := &http.Server{
		RateLimiter: rl,
	}

	// Allow app to listen for OS Interrupts and SIGTERMS.
	serverErrors := make(chan error, 1)
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	go func() {
		serverErrors <- s.Run()
	}()

	// Handling for server errors and OS signals.
	select {
	case err := <-serverErrors:
		log.Fatalf("error starting server: %v", err.Error())
	case <-osSignals:
		log.Println("starting server shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := s.Shutdown(ctx)
		if err != nil {
			log.Fatalf("error trying to shutdown http server: %v", err.Error())
		}
	}
}
