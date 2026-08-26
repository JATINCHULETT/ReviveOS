package runner

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/utils/outbox"
	"github.com/reviveos/utils/queue"
	"github.com/reviveos/worker/internal/tasks"
)

// StartEmbeddedWorker starts the Outbox Relay and Asynq Worker background process.
func StartEmbeddedWorker(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Initializing Outbox Relay & Asynq Background Worker...")
	if err := queue.PingRedis(ctx); err != nil {
		log.Printf("WARN: Background worker could not reach Redis: %v. Background tasks will wait.", err)
		return
	}

	asynqClient := queue.NewClient()
	defer asynqClient.Close()

	// Start Outbox Relay
	relay := outbox.NewRelay(pool, asynqClient, 25)
	go relay.Start(ctx, 250*time.Millisecond)
	log.Println("Outbox Relay active (polling interval: 250ms).")

	// Start Asynq Server
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

	go func() {
		<-ctx.Done()
		srv.Shutdown()
	}()

	log.Println("Asynq Worker listening for background recovery tasks...")
	if err := srv.Run(mux); err != nil {
		log.Printf("WARN: Asynq server stopped: %v", err)
	}
}
