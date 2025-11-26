package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/mini-hotel-aggregator/internal/app"
)

func main() {
	ctx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	//Create AppConfig will all initialization
	appConfig := app.SetAppConfig()

	srv := &http.Server{
		Addr:    addr,
		Handler: appConfig.Router,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	// graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %v, initiating graceful shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
		// Cancel root context so ALL goroutines & requests stop
		log.Println("Shutdown done...Cancelling all goroutines which are still running")
		rootCancel()
		close(idleConnsClosed)
	}()

	log.Printf("starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	<-idleConnsClosed
	log.Printf("server stopped")
}
