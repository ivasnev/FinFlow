package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"

	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmtransport "github.com/ivasnev/FinFlow/ff-tvm/pkg/transport"
)

// Client представляет клиент для взаимодействия с сервисом идентификации пользователей
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// RegisterUserRequest представляет запрос на регистрацию пользователя
type RegisterUserRequest struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// NewClient создает новый клиент для взаимодействия с сервисом идентификации пользователей
func NewClient(baseURL string, fromServiceID, toServiceID int, tvmClient *tvmclient.TVMClient) *Client {
	tvmTransport := tvmtransport.NewTVMTransport(tvmClient, http.DefaultTransport, fromServiceID, toServiceID)

	httpClient := &http.Client{
		Transport: tvmTransport,
		Timeout:   10 * time.Second,
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// RegisterUser регистрирует нового пользователя
func (c *Client) RegisterUser(ctx context.Context, reqBody *RegisterUserRequest) (*dto.UserDTO, error) {

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/internal/users/register", c.baseURL),
		bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("ошибка сервера: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("ошибка сервера: %s", errResp.Error)
	}

	// Парсим ответ
	var userFromResponse *dto.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&userFromResponse); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	// Возвращаем данные пользователя
	return userFromResponse, nil
}
