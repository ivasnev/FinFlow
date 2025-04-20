package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmtransport "github.com/ivasnev/FinFlow/ff-tvm/pkg/transport"
)

// Client - клиент для взаимодействия с сервисом Split
type Client struct {
	baseURL    string
	tvmClient  *tvmclient.TVMClient
	from       int
	to         int
	httpClient *http.Client
}

// NewClient - создает новый экземпляр клиента
func NewClient(baseURL string, tvmClient *tvmclient.TVMClient, fromServiceID int, toServiceID int) *Client {
	tvmTransport := tvmtransport.NewTVMTransport(tvmClient, http.DefaultTransport, fromServiceID, toServiceID)

	httpClient := &http.Client{
		Transport: tvmTransport,
		Timeout:   10 * time.Second,
	}
	return &Client{
		baseURL:    baseURL,
		tvmClient:  tvmClient,
		from:       fromServiceID,
		to:         toServiceID,
		httpClient: httpClient,
	}
}

// Пример метода для проверки состояния сервиса
func (c *Client) Health() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/health", c.baseURL), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Status == "ok", nil
}

// HealthInternal - проверка с использованием TVM
func (c *Client) HealthInternal() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/internal/split/health", c.baseURL), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Status == "ok", nil
}

// ApiRequest выполняет запрос к API с TVM тикетом
func (c *Client) ApiRequest(method, path string, payload any) (*http.Response, error) {

	var requestBody *bytes.Buffer
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	} else {
		requestBody = bytes.NewBuffer(nil)
	}

	// Формируем URL запроса
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	// Создаем запрос
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем заголовки
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}
