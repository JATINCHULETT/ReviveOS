package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reviveos/utils/queue"
)

type QueueStats struct {
	Queue       string        `json:"queue"`
	Size        int           `json:"size"`
	MemoryUsage int64         `json:"memory_usage_bytes"`
	Latency     time.Duration `json:"latency"`
	Active      int           `json:"active"`
	Pending     int           `json:"pending"`
	Scheduled   int           `json:"scheduled"`
	Retry       int           `json:"retry"`
	Archived    int           `json:"archived"`
	Completed   int           `json:"completed"`
	Paused      bool          `json:"paused"`
}

type ServerInfo struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	PID         int       `json:"pid"`
	Concurrency int       `json:"concurrency"`
	Queues      []string  `json:"queues"`
	Started     time.Time `json:"started"`
	Status      string    `json:"status"`
}

type SystemQueuesResponse struct {
	RedisStatus string       `json:"redis_status"`
	Queues      []QueueStats `json:"queues"`
	Servers     []ServerInfo `json:"servers"`
	Timestamp   time.Time    `json:"timestamp"`
}

// SystemQueuesHandler returns actual queue and server stats from Redis/Asynq.
func SystemQueuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	ctx := r.Context()
	redisErr := queue.PingRedis(ctx)
	redisStatus := "connected"
	if redisErr != nil {
		redisStatus = "unreachable"
		log.Printf("ERROR: system queues redis ping failed: %v", redisErr)
	}

	inspector := queue.NewInspector()
	defer inspector.Close()

	queuesList, err := inspector.Queues()
	if err != nil && redisStatus == "connected" {
		log.Printf("ERROR: failed to fetch queues from asynq: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to inspect queues", err.Error())
		return
	}

	// If no queues exist yet, default queue is at least inspectable
	if len(queuesList) == 0 {
		queuesList = []string{"default"}
	}

	var statsList []QueueStats
	for _, qName := range queuesList {
		qInfo, err := inspector.GetQueueInfo(qName)
		if err != nil {
			statsList = append(statsList, QueueStats{
				Queue: qName,
			})
			continue
		}

		statsList = append(statsList, QueueStats{
			Queue:       qInfo.Queue,
			Size:        qInfo.Size,
			MemoryUsage: qInfo.MemoryUsage,
			Latency:     qInfo.Latency,
			Active:      qInfo.Active,
			Pending:     qInfo.Pending,
			Scheduled:   qInfo.Scheduled,
			Retry:       qInfo.Retry,
			Archived:    qInfo.Archived,
			Completed:   qInfo.Completed,
			Paused:      qInfo.Paused,
		})
	}

	var serversList []ServerInfo
	servers, err := inspector.Servers()
	if err == nil {
		for _, s := range servers {
			var qNames []string
			for qName := range s.Queues {
				qNames = append(qNames, qName)
			}
			serversList = append(serversList, ServerInfo{
				ID:          s.ID,
				Host:        s.Host,
				PID:         s.PID,
				Concurrency: s.Concurrency,
				Queues:      qNames,
				Started:     s.Started,
				Status:      s.Status,
			})
		}
	}

	resp := SystemQueuesResponse{
		RedisStatus: redisStatus,
		Queues:      statsList,
		Servers:     serversList,
		Timestamp:   time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type ComponentStatus struct {
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // HEALTHY, DEGRADED, UNHEALTHY, AVAILABLE, UNAVAILABLE
	LatencyMs   int64                  `json:"latency_ms"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

type SystemHealthResponse struct {
	OverallStatus string            `json:"overall_status"`
	Components    []ComponentStatus `json:"components"`
	Timestamp     time.Time         `json:"timestamp"`
}

// SystemHealthHandler returns live operational health across all ReviveOS infrastructure dependencies.
func SystemHealthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		var components []ComponentStatus
		allHealthy := true

		// 1. API
		components = append(components, ComponentStatus{
			Name:      "API Server",
			Status:    "HEALTHY",
			LatencyMs: 0,
			Message:   "HTTP API accepting requests",
		})

		// 2. PostgreSQL
		pgStart := time.Now()
		var pgStatus = "HEALTHY"
		var pgMsg = "Connected and operational"
		if pool == nil {
			pgStatus = "UNHEALTHY"
			pgMsg = "Pool is nil"
			allHealthy = false
		} else if err := pool.Ping(ctx); err != nil {
			pgStatus = "UNHEALTHY"
			pgMsg = err.Error()
			allHealthy = false
		}
		pgLatency := time.Since(pgStart).Milliseconds()
		components = append(components, ComponentStatus{
			Name:      "PostgreSQL",
			Status:    pgStatus,
			LatencyMs: pgLatency,
			Message:   pgMsg,
		})

		// 3. Redis
		redisStart := time.Now()
		var redisStatus = "HEALTHY"
		var redisMsg = "Connected to Redis instance"
		if err := queue.PingRedis(ctx); err != nil {
			redisStatus = "UNHEALTHY"
			redisMsg = err.Error()
			allHealthy = false
		}
		redisLatency := time.Since(redisStart).Milliseconds()
		components = append(components, ComponentStatus{
			Name:      "Redis",
			Status:    redisStatus,
			LatencyMs: redisLatency,
			Message:   redisMsg,
		})

		// 4. Asynq Queue / Workers
		asynqStart := time.Now()
		var asynqStatus = "HEALTHY"
		var asynqMsg = "Asynq queue worker pool active"
		inspector := queue.NewInspector()
		servers, sErr := inspector.Servers()
		inspector.Close()
		if sErr != nil {
			asynqStatus = "DEGRADED"
			asynqMsg = sErr.Error()
		} else if len(servers) == 0 {
			asynqStatus = "DEGRADED"
			asynqMsg = "No active Asynq worker servers running"
		}
		components = append(components, ComponentStatus{
			Name:      "Asynq Worker",
			Status:    asynqStatus,
			LatencyMs: time.Since(asynqStart).Milliseconds(),
			Message:   asynqMsg,
			Details: map[string]interface{}{
				"active_servers": len(servers),
			},
		})

		// 5. Ollama & DeepSeek
		ollamaURL := os.Getenv("OLLAMA_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
		ollamaStart := time.Now()
		ollamaStatus := "HEALTHY"
		ollamaMsg := "Ollama service online"
		deepseekStatus := "AVAILABLE"
		deepseekMsg := "deepseek-r1:1.5b loaded"

		client := &http.Client{Timeout: 2 * time.Second}
		req, _ := http.NewRequestWithContext(ctx, "GET", ollamaURL+"/api/tags", nil)
		resp, err := client.Do(req)
		if err != nil {
			ollamaStatus = "UNHEALTHY"
			ollamaMsg = err.Error()
			deepseekStatus = "UNAVAILABLE"
			deepseekMsg = "Cannot verify model (Ollama unreachable)"
			allHealthy = false
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ollamaStatus = "UNHEALTHY"
				ollamaMsg = fmt.Sprintf("HTTP status %d", resp.StatusCode)
				deepseekStatus = "UNAVAILABLE"
			} else {
				var tagsResp struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err == nil {
					foundModel := false
					for _, m := range tagsResp.Models {
						if m.Name == "deepseek-r1:1.5b" || m.Name == "deepseek-r1:1.5b-latest" {
							foundModel = true
							break
						}
					}
					if !foundModel {
						deepseekStatus = "UNAVAILABLE"
						deepseekMsg = "deepseek-r1:1.5b model not pulled"
					}
				}
			}
		}

		components = append(components, ComponentStatus{
			Name:      "Ollama",
			Status:    ollamaStatus,
			LatencyMs: time.Since(ollamaStart).Milliseconds(),
			Message:   ollamaMsg,
		})

		components = append(components, ComponentStatus{
			Name:      "DeepSeek-R1 (1.5B)",
			Status:    deepseekStatus,
			LatencyMs: time.Since(ollamaStart).Milliseconds(),
			Message:   deepseekMsg,
		})

		// 6. Payment Provider
		provType := os.Getenv("PAYMENT_PROVIDER")
		if provType == "" {
			provType = "local"
		}
		components = append(components, ComponentStatus{
			Name:      "Payment Provider",
			Status:    "HEALTHY",
			LatencyMs: 0,
			Message:   fmt.Sprintf("Active provider: %s", provType),
		})

		overall := "HEALTHY"
		if !allHealthy {
			overall = "DEGRADED"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SystemHealthResponse{
			OverallStatus: overall,
			Components:    components,
			Timestamp:     time.Now().UTC(),
		})
	}
}

// RootHandler handles requests to "/" providing service metadata and API endpoints directory.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	response := map[string]interface{}{
		"status":      "online",
		"service":     "ReviveOS Autonomous Payment Recovery API",
		"version":     "1.0.0",
		"environment": os.Getenv("APP_ENV"),
		"timestamp":   time.Now().UTC(),
		"endpoints": map[string]string{
			"health":             "/health",
			"system_health":      "/system/health",
			"system_queues":      "/system/queues",
			"workflows":          "/workflows",
			"analytics_overview": "/analytics/overview",
			"razorpay_webhook":   "/webhooks/razorpay",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}


