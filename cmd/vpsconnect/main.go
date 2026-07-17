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

	"github.com/Fregres/VPSconnect/internal/httpapi"
	"github.com/Fregres/VPSconnect/internal/metrics"
	"github.com/Fregres/VPSconnect/internal/storage/postgres"
)

func main() {

	token, exists := os.LookupEnv("VPSCONNECT_TOKEN")
	databaseURL := os.Getenv("DATABASE_URL")

	if strings.TrimSpace(databaseURL) == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if !exists || strings.TrimSpace(token) == "" {
		log.Fatal("VPSCONNECT_TOKEN is required")
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startupCancel()

	storage, err := postgres.New(startupCtx, databaseURL)
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer storage.Close()

	log.Println("Connected to PostgreSQL")
	status, err := metrics.CollectStatus()
	if err != nil {
		log.Fatalf("collect initial status: %v", err)
	}

	if err := storage.SaveMetric(startupCtx, status); err != nil {
		log.Fatalf("save initial metric: %v", err)
	}

	log.Println("Initial metric saved")
	srv := httpapi.NewServer(token)

	const address = "127.0.0.1:6767"
	server := &http.Server{
		Addr:              address,
		Handler:           srv.Handler(),
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
