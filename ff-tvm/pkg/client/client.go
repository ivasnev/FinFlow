package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TVMClient struct {
	baseURL    string
	httpClient *http.Client
	secret     string
}

func NewTVMClient(baseURL string, secret string) *TVMClient {
	return &TVMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 5,
		},
		secret: secret,
	}
}

func (c *TVMClient) GetPublicKey(serviceID int) (string, error) {
	url := fmt.Sprintf("%s/service/%d/key", c.baseURL, serviceID)

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

func (c *TVMClient) GenerateTicket(from, to int) (string, error) {
	url := fmt.Sprintf("%s/ticket", c.baseURL)

	reqBody := struct {
		From   int    `json:"from"`
		To     int    `json:"to"`
		Secret string `json:"secret"`
	}{
		From:   from,
		To:     to,
		Secret: c.secret,
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Ticket string `json:"ticket"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Ticket, nil
}
