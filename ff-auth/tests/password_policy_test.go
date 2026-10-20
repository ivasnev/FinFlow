package tests

import (
	"context"
	"net/http"
	"strings"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TestRegister_PasswordPolicy проверяет HTTP-отказ до создания пользователя и сессии.
func (s *RegisterSuite) TestRegister_PasswordPolicy() {
	ctx := context.Background()
	for _, password := range []string{"x", "", "1234567", strings.Repeat("я", 4), strings.Repeat("a", 73)} {
		resp, err := s.APIClient.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: openapi_types.Email("policy@example.com"), Nickname: "policy", Password: password})
		s.Require().NoError(err)
		s.Require().Equal(400, resp.StatusCode())
	}
	previous := s.Config.Auth.PasswordMinLength
	s.Config.Auth.PasswordMinLength = 12
	defer func() { s.Config.Auth.PasswordMinLength = previous }()
	resp, err := s.APIClient.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: openapi_types.Email("policy@example.com"), Nickname: "policy", Password: "12345678901"})
	s.Require().NoError(err)
	s.Require().Equal(400, resp.StatusCode())
	for _, table := range []string{"users", "sessions"} {
		var count int64
		s.Require().NoError(s.GetDB().Table(table).Count(&count).Error)
		s.Equal(int64(0), count, table)
	}
	s.MockServer.Expect(http.MethodPost, "/api/v1/internal/users/register").Return("ff_id_service/register_user_response_success.json").HTTPCode(201)
	valid, err := s.APIClient.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: openapi_types.Email("policy@example.com"), Nickname: "policy", Password: "123456789012"})
	s.Require().NoError(err)
	s.Require().Equal(201, valid.StatusCode())
}

// TestUpdateUser_PasswordPolicy проверяет, что неверный новый пароль не меняет профиль или действующий пароль.
func (s *UsersSuite) TestUpdateUser_PasswordPolicy() {
	ctx := context.Background()
	s.MockServer.Expect(http.MethodPost, "/api/v1/internal/users/register").Return("ff_id_service/register_user_response_success.json").HTTPCode(201)
	registered, err := s.APIClient.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: openapi_types.Email("change@example.com"), Nickname: "change", Password: "password123"})
	s.Require().NoError(err)
	s.Require().Equal(201, registered.StatusCode())
	s.Require().NotNil(registered.JSON201)
	authorize := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registered.JSON201.AccessToken)
		return nil
	}
	var original struct{ PasswordHash, Email, Nickname string }
	s.Require().NoError(s.GetDB().Table("users").Where("id = ?", registered.JSON201.User.Id).Take(&original).Error)
	for _, password := range []string{"x", "", "1234567", strings.Repeat("я", 4), strings.Repeat("a", 73)} {
		email := openapi_types.Email("unexpected@example.com")
		nickname := "unexpected"
		resp, err := s.APIClient.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Password: &password, Email: &email, Nickname: &nickname}, authorize)
		s.Require().NoError(err)
		s.Require().Equal(400, resp.StatusCode())
		var current struct{ PasswordHash, Email, Nickname string }
		s.Require().NoError(s.GetDB().Table("users").Where("id = ?", registered.JSON201.User.Id).Take(&current).Error)
		s.Equal(original, current)
	}
	// Сценарий: пароль ровно из восьми символов сохраняется и подходит для входа.
	newPassword := "12345678"
	updated, err := s.APIClient.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Password: &newPassword}, authorize)
	s.Require().NoError(err)
	s.Require().Equal(200, updated.StatusCode())
	login, err := s.APIClient.LoginWithResponse(ctx, api.LoginJSONRequestBody{Login: "change@example.com", Password: newPassword})
	s.Require().NoError(err)
	s.Require().Equal(200, login.StatusCode())
}
