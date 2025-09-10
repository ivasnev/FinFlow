package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
)

type FilesClient interface {
	UploadFile(ctx context.Context, fileData []byte, filename string) (string, error)
	DeleteFile(ctx context.Context, fileID string) error
}

type filesClient struct {
	baseURL    string
	httpClient *http.Client
	tvmClient  tvmclient.Client
}

func NewFilesClient(baseURL string, tvmClient tvmclient.Client) FilesClient {
	return &filesClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		tvmClient:  tvmClient,
	}
}

func (c *filesClient) UploadFile(ctx context.Context, fileData []byte, filename string) (string, error) {
	// Получаем сервисный тикет
	ticket, err := c.tvmClient.GetServiceTicket(ctx, "ff-files")
	if err != nil {
		return "", fmt.Errorf("failed to get service ticket: %w", err)
	}

	// Создаем multipart форму
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := io.Copy(part, bytes.NewReader(fileData)); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/upload", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+ticket)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Парсим ответ
	var response struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.ID, nil
}

func (c *filesClient) DeleteFile(ctx context.Context, fileID string) error {
	// Получаем сервисный тикет
	ticket, err := c.tvmClient.GetServiceTicket(ctx, "ff-files")
	if err != nil {
		return fmt.Errorf("failed to get service ticket: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/file/%s", c.baseURL, fileID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ticket)

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
} 