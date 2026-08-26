package queue

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// GetRedisOpt parses REDIS_URL from environment or returns default options
func GetRedisOpt() asynq.RedisConnOpt {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	redisURL = strings.ReplaceAll(redisURL, "\r", "")
	redisURL = strings.ReplaceAll(redisURL, "\n", "")
	redisURL = strings.ReplaceAll(redisURL, "\t", "")

	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		// Fallback to simple host:port if parsing failed
		addr := strings.TrimPrefix(redisURL, "redis://")
		if strings.Contains(addr, "/") {
			addr = strings.Split(addr, "/")[0]
		}
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		return asynq.RedisClientOpt{
			Addr: addr,
		}
	}

	return opt
}

// PingRedis checks connectivity to Redis
func PingRedis(ctx context.Context) error {
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	redisURL = strings.ReplaceAll(redisURL, "\r", "")
	redisURL = strings.ReplaceAll(redisURL, "\n", "")
	redisURL = strings.ReplaceAll(redisURL, "\t", "")

	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		opt = &redis.Options{
			Addr: "127.0.0.1:6379",
		}
	}

	rdb := redis.NewClient(opt)
	defer rdb.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// NewClient creates a new Asynq Client
func NewClient() *asynq.Client {
	return asynq.NewClient(GetRedisOpt())
}

// NewInspector creates a new Asynq Inspector
func NewInspector() *asynq.Inspector {
	return asynq.NewInspector(GetRedisOpt())
}
