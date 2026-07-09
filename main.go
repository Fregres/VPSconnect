package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	token, exists := os.LookupEnv("VPSCONNECT_TOKEN")

	if !exists || strings.TrimSpace(token) == "" {
		log.Fatal("VPSCONNECT_TOKEN is required")
	}

	srv := &Server{token: token}
	const address = "127.0.0.1:6767"
	server := &http.Server{
		Addr:              address,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("server started on %s", address)
	errChan := make(chan error, 1)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, ShutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ShutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
		log.Println("Graceful shutdown complete")
	case err := <-errChan:
		log.Fatalf("ListenAndServe error:%v", err)
	}

}
