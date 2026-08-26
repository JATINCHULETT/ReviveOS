package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/reviveos/api/handlers"
	"github.com/reviveos/utils/db"
	"github.com/reviveos/worker/runner"
)

func main() {
	log.Println("Starting ReviveOS API Server...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Database connection pool initialized.")

	// Server run context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Start Embedded Outbox Relay & Asynq Worker unless explicitly disabled
	if os.Getenv("ENABLE_EMBEDDED_WORKER") != "false" {
		go runner.StartEmbeddedWorker(serverCtx, pool)
	}

	handler := handlers.RegisterRoutes(pool)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 5 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 5*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	log.Printf("ReviveOS API listening on port %s (HTTP)", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: Server error: %v", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	log.Println("ReviveOS API Server stopped cleanly.")
}
