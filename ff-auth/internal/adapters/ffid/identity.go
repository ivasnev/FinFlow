package ffid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ErrNotFound отличает отсутствующий логин от недоступности источника идентификаторов.
var ErrNotFound = errors.New("пользователь не найден")

// ErrConflict обозначает занятые идентификаторы.
var ErrConflict = errors.New("указанный email или никнейм уже используется")

// ErrUnavailable запрещает вход по устаревшим локальным копиям при отказе ff-id.
var ErrUnavailable = errors.New("источник идентификаторов недоступен")

const (
	LookupLogin    = "login"
	LookupEmail    = "email"
	LookupNickname = "nickname"
)

// FindIdentity находит актуальный профиль по явно заданному типу идентификатора.
func (a *Adapter) FindIdentity(ctx context.Context, login string, kind string) (*UserDTO, error) {
	searchKind := api.LookupIdentityParamsKind(kind)
	return a.lookupIdentity(ctx, api.LookupIdentityParams{Login: &login, Kind: &searchKind})
}

// GetIdentity получает актуальный профиль по стабильному ID учётных данных.
func (a *Adapter) GetIdentity(ctx context.Context, id int64) (*UserDTO, error) {
	return a.lookupIdentity(ctx, api.LookupIdentityParams{UserId: &id})
}

// lookupIdentity проверяет статус и наличие обязательного тела ответа.
func (a *Adapter) lookupIdentity(ctx context.Context, params api.LookupIdentityParams) (*UserDTO, error) {
	resp, err := a.client.LookupIdentityWithResponse(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil || resp.JSON200.Id <= 0 {
		return nil, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode())
	}
	mismatch := params.UserId != nil && resp.JSON200.Id != *params.UserId
	if params.Login != nil {
		switch {
		case params.Kind != nil && string(*params.Kind) == LookupEmail:
			mismatch = !strings.EqualFold(string(resp.JSON200.Email), *params.Login)
		case params.Kind != nil && string(*params.Kind) == LookupNickname:
			mismatch = resp.JSON200.Nickname != *params.Login
		default:
			mismatch = !strings.EqualFold(string(resp.JSON200.Email), *params.Login) && resp.JSON200.Nickname != *params.Login
		}
	}
	if mismatch {
		return nil, fmt.Errorf("%w: профиль не соответствует запросу", ErrUnavailable)
	}
	return convertUserDTO(resp.JSON200), nil
}

// UpdateIdentity направляет изменение в ff-id без второй записи email и nickname в auth.
func (a *Adapter) UpdateIdentity(ctx context.Context, id int64, email, nickname *string) (*UserDTO, error) {
	req := api.UpdateIdentityJSONRequestBody{Nickname: nickname}
	if email != nil {
		value := openapi_types.Email(*email)
		req.Email = &value
	}
	resp, err := a.client.UpdateIdentityWithResponse(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if resp.StatusCode() == http.StatusConflict {
		return nil, ErrConflict
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil || resp.JSON200.Id != id {
		return nil, fmt.Errorf("%w: изменение не подтверждено (status %d)", ErrUnavailable, resp.StatusCode())
	}
	return convertUserDTO(resp.JSON200), nil
}
