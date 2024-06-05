package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebhookData struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Priority string `json:"priority"`
	Source   string `json:"source"`
}

func Send(webhookURL string, data WebhookData) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Make an HTTP POST request to the webhook URL with the JSON payload
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
