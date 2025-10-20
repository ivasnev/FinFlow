package ffid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
)

// Adapter - адаптер для работы с ff-id сервисом через сгенерированный клиент
type Adapter struct {
	client *api.ClientWithResponses
}

// NewAdapter создает новый адаптер для ff-id клиента
func NewAdapter(baseURL string, httpClient *http.Client) (*Adapter, error) {
	client, err := api.NewClientWithResponses(
		baseURL,
		api.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}
	return &Adapter{client: client}, nil
}

// GetUsersByIDs получает информацию о пользователях по их ID
func (a *Adapter) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]UserDTO, error) {
	if len(userIDs) == 0 {
		return []UserDTO{}, nil
	}

	params := api.GetUsersByIdsParams{
		UserId: userIDs,
	}

	resp, err := a.client.GetUsersByIdsWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	// Обработка различных статусов
	if resp.StatusCode() != http.StatusOK {
		if resp.JSON400 != nil {
			return nil, fmt.Errorf("ошибка сервера: %s", resp.JSON400.Error)
		}
		if resp.JSON500 != nil {
			return nil, fmt.Errorf("ошибка сервера: %s", resp.JSON500.Error)
		}
		return nil, fmt.Errorf("неожиданный статус: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return []UserDTO{}, nil
	}

	// Конвертация из API типов в адаптерные типы
	users := make([]UserDTO, 0, len(*resp.JSON200))
	for _, apiUser := range *resp.JSON200 {
		users = append(users, convertUserDTO(&apiUser))
	}

	return users, nil
}

// GetUserByID получает информацию о пользователе по его ID
func (a *Adapter) GetUserByID(ctx context.Context, userID int64) (*UserDTO, error) {
	users, err := a.GetUsersByIDs(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("пользователь с ID %d не найден", userID)
	}

	return &users[0], nil
}

// convertUserDTO конвертирует API UserDTO в адаптерный UserDTO
func convertUserDTO(apiUser *api.UserDTO) UserDTO {
	var avatarID *string
	if apiUser.AvatarId != nil {
		avatarStr := apiUser.AvatarId.String()
		avatarID = &avatarStr
	}

	return UserDTO{
		ID:        apiUser.Id,
		Email:     string(apiUser.Email),
		Nickname:  apiUser.Nickname,
		Name:      apiUser.Name,
		Phone:     apiUser.Phone,
		AvatarID:  avatarID,
		Birthdate: apiUser.Birthdate,
		CreatedAt: apiUser.CreatedAt,
		UpdatedAt: apiUser.UpdatedAt,
	}
}
