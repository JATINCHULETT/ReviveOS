package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": "deepseek-r1:1.5b",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Payment failed with INSUFFICIENT_FUNDS. Output a valid JSON recovery recommendation with keys: diagnosis, recoverability, recommended_action (e.g. DELAYED_RETRY), recommended_delay_hours (e.g. 24), reason, confidence.",
			},
		},
		"stream": false,
	})

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	json.Unmarshal(body, &data)
	fmt.Printf("Content: %s\n", data.Message.Content)
}
