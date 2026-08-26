package db

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect returns a connection pool to the database
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	connStr := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	connStr = strings.ReplaceAll(connStr, "\r", "")
	connStr = strings.ReplaceAll(connStr, "\n", "")
	connStr = strings.ReplaceAll(connStr, "\t", "")

	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/reviveos?sslmode=disable"
	}

	// Try direct parse first
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		// If failed due to unescaped characters in password, attempt auto-escaping
		if strings.HasPrefix(connStr, "postgresql://") || strings.HasPrefix(connStr, "postgres://") {
			parts := strings.SplitN(connStr, "://", 2)
			if len(parts) == 2 {
				atIndex := strings.LastIndex(parts[1], "@")
				if atIndex != -1 {
					userInfo := parts[1][:atIndex]
					hostAndDb := parts[1][atIndex+1:]
					userPass := strings.SplitN(userInfo, ":", 2)
					if len(userPass) == 2 {
						escapedPass := url.QueryEscape(userPass[1])
						escapedPass = strings.ReplaceAll(escapedPass, "+", "%20")
						reconstructed := fmt.Sprintf("%s://%s:%s@%s", parts[0], userPass[0], escapedPass, hostAndDb)
						config, err = pgxpool.ParseConfig(reconstructed)
					}
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
