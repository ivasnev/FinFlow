package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TVMClient interface {
	ValidateTicket(ticket string) (string, error)
}

type tvmClient struct {
	baseURL     string
	serviceID   string
	serviceKey  string
	httpClient  *http.Client
}

func NewTVMClient(baseURL, serviceID, serviceKey string) TVMClient {
	return &tvmClient{
		baseURL:    baseURL,
		serviceID:  serviceID,
		serviceKey: serviceKey,
		httpClient: &http.Client{},
	}
}

func (c *tvmClient) ValidateTicket(ticket string) (string, error) {
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