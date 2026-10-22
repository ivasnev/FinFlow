package ffid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

// RegisterUser регистрирует нового пользователя в сервисе ff-id
func (a *Adapter) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*UserDTO, error) {
	// Конвертация типов из адаптера в API типы
	apiReq := api.ServiceRegisterUserRequest{
		UserId:   req.UserID,
		Email:    openapi_types.Email(req.Email),
		Nickname: req.Nickname,
		Name:     req.Name,
		Phone:    req.Phone,
	}

	resp, err := a.client.RegisterUserFromServiceWithResponse(ctx, apiReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	if resp.StatusCode() == http.StatusConflict {
		return nil, ErrConflict
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode())
	}

	if resp.JSON201 == nil || resp.JSON201.Id != req.UserID {
		return nil, fmt.Errorf("%w: некорректный ID созданного профиля", ErrUnavailable)
	}

	// Конвертация из API типов в адаптерные типы
	return convertUserDTO(resp.JSON201), nil
}
