package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/reviveos/simulator/generator"
)

func main() {
	count := flag.Int("count", 50, "Number of synthetic events to generate and send")
	apiURL := flag.String("api", "http://localhost:8080/simulator/failure", "API endpoint to send events")
	flag.Parse()

	log.Printf("Generating %d synthetic payment failure events...", *count)

	gen := generator.New()
	events := gen.GenerateBatch(*count)

	client := &http.Client{Timeout: 10 * time.Second}
	successes := 0

	for i, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			log.Printf("Failed to serialize event %d: %v", i, err)
			continue
		}

		resp, err := client.Post(*apiURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Failed to send event %d: %v", i, err)
			continue
		}

		if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK {
			successes++
		} else {
			log.Printf("Unexpected status code %d for event %d", resp.StatusCode, i)
		}
		resp.Body.Close()
		
		// Small delay to prevent overwhelming the local API during batch testing
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Printf("Successfully sent %d/%d synthetic events to ReviveOS API.\n", successes, *count)
}
