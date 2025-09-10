package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client интерфейс клиента TVM
type Client interface {
	GetServiceTicket(ctx context.Context, targetService string) (string, error)
	ValidateTicket(ticket string) (string, error)
}

type client struct {
	baseURL     string
	serviceID   string
	serviceKey  string
	httpClient  *http.Client
}

// Config конфигурация клиента TVM
type Config struct {
	BaseURL    string
	ServiceID  string
	ServiceKey string
}

// NewClient создает новый клиент TVM
func NewClient(cfg Config) Client {
	return &client{
		baseURL:    cfg.BaseURL,
		serviceID:  cfg.ServiceID,
		serviceKey: cfg.ServiceKey,
		httpClient: &http.Client{},
	}
}

func (c *client) GetServiceTicket(ctx context.Context, targetService string) (string, error) {
	// Подготавливаем запрос
	reqBody := struct {
		ServiceID     string `json:"service_id"`
		TargetService string `json:"target_service"`
	}{
		ServiceID:     c.serviceID,
		TargetService: targetService,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/ticket", strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", c.serviceKey)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Парсим ответ
	var response struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Ticket, nil
}

func (c *client) ValidateTicket(ticket string) (string, error) {
	// Подготавливаем запрос
	reqBody := struct {
		Ticket string `json:"ticket"`
	}{
		Ticket: ticket,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequest("POST", c.baseURL+"/validate", strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Key", c.serviceKey)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Парсим ответ
	var response struct {
		ServiceID string `json:"service_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.ServiceID, nil
} 