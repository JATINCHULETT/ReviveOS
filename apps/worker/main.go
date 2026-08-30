package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/reviveos/utils/db"
	"github.com/reviveos/utils/env"
	"github.com/reviveos/utils/outbox"
	"github.com/reviveos/utils/queue"
	"github.com/reviveos/worker/internal/tasks"
)

func main() {
	env.Load()
	log.Println("Starting ReviveOS Worker & Outbox Relay...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Database Connection
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Database connection pool initialized.")

	// 2. Redis Check & Asynq Client
	if err := queue.PingRedis(context.Background()); err != nil {
		log.Fatalf("FATAL: Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connectivity verified.")

	asynqClient := queue.NewClient()
	defer asynqClient.Close()

	// 3. Outbox Relay
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()

	relay := outbox.NewRelay(pool, asynqClient, 25)
	go relay.Start(relayCtx, 250*time.Millisecond)
	log.Println("Transactional Outbox Relay started (polling interval: 250ms).")

	// 4. Asynq Server
	srv := asynq.NewServer(
		queue.GetRedisOpt(),
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			LogLevel: asynq.InfoLevel,
		},
	)

	mux := asynq.NewServeMux()
	taskHandler := tasks.NewTaskHandler(pool)
	taskHandler.RegisterRoutes(mux)

	// Graceful shutdown handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("Shutdown signal received, shutting down worker and relay...")
		relayCancel()
		srv.Shutdown()
	}()

	log.Println("ReviveOS Asynq Worker listening for background tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("FATAL: Asynq server run error: %v", err)
	}

	log.Println("ReviveOS Worker stopped cleanly.")
}
