package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TVMClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewTVMClient(baseURL string) *TVMClient {
	return &TVMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

func (c *TVMClient) GetPublicKey(serviceID int64) (string, error) {
	url := fmt.Sprintf("%s/service/%d/pub_key", c.baseURL, serviceID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		PublicKey string `json:"public_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.PublicKey, nil
}

func (c *TVMClient) GenerateTicket(from, to int64) (string, error) {
	url := fmt.Sprintf("%s/ticket", c.baseURL)

	reqBody := struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
	}{
		From: from,
		To:   to,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to generate ticket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Ticket string `json:"ticket"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Ticket, nil
}
