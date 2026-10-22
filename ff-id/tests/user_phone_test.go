package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// profileWrite выполняет HTTP-запись с явно заданным тестовым пользователем.
func (s *UserSuite) profileWrite(actor int64, method, path string, body map[string]any) *httptest.ResponseRecorder {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(auth.UserContextKey()), auth.UserData{UserID: actor})
		c.Next()
	})
	router.PATCH("/users/me", s.Container.ServerHandler.UpdateUser)
	router.POST("/users/register", s.Container.ServerHandler.RegisterUser)
	data, err := json.Marshal(body)
	s.Require().NoError(err)
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

// TestPhone_OptionalAndConflict проверяет NULL, сохранение номера и атомарный конфликт через HTTP и PostgreSQL.
func (s *UserSuite) TestPhone_OptionalAndConflict() {
	s.createTestUser(1, "one@example.com", "one", "One")
	s.createTestUser(2, "two@example.com", "two", "Two")
	for _, actor := range []int64{1, 2} {
		for _, phone := range []string{"", " \t\n", "\u00a0"} {
			response := s.profileWrite(actor, http.MethodPatch, "/users/me", map[string]any{"phone": phone, "name": "Updated"})
			s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
			var user map[string]any
			s.Require().NoError(json.Unmarshal(response.Body.Bytes(), &user))
			s.NotContains(user, "phone")
		}
	}
	var missing int64
	s.Require().NoError(s.GetDB().Table("users").Where("phone IS NULL").Count(&missing).Error)
	s.Equal(int64(2), missing)
	response := s.profileWrite(1, http.MethodPatch, "/users/me", map[string]any{"phone": "+70000000012"})
	s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
	response = s.profileWrite(1, http.MethodPatch, "/users/me", map[string]any{"name": "Kept phone"})
	s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
	user, err := s.Container.UserService.GetUserByID(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().NotNil(user.Phone)
	s.Equal("+70000000012", *user.Phone)
	response = s.profileWrite(2, http.MethodPatch, "/users/me", map[string]any{"phone": "+70000000012", "name": "Must not persist"})
	s.Equal(http.StatusConflict, response.Code, response.Body.String())
	s.Contains(response.Body.String(), service.ErrPhoneConflict.Error())
	s.NotContains(response.Body.String(), "email")
	user, err = s.Container.UserService.GetUserByID(s.Ctx, 2)
	s.Require().NoError(err)
	s.Nil(user.Phone)
	s.Equal("Updated", *user.Name)
	for _, body := range []map[string]any{{"email": "one@example.com"}, {"nickname": "one"}} {
		response = s.profileWrite(2, http.MethodPatch, "/users/me", body)
		if _, emailChange := body["email"]; emailChange {
			s.Equal(http.StatusBadRequest, response.Code, response.Body.String())
			s.Contains(response.Body.String(), "подтверждение")
		} else {
			s.Equal(http.StatusConflict, response.Code, response.Body.String())
			s.Contains(response.Body.String(), service.ErrIdentityConflict.Error())
		}
	}
	response = s.profileWrite(1, http.MethodPatch, "/users/me", map[string]any{"phone": ""})
	s.Require().Equal(http.StatusOK, response.Code, response.Body.String())
	user, err = s.Container.UserService.GetUserByID(s.Ctx, 1)
	s.Require().NoError(err)
	s.Nil(user.Phone)
}

// TestPhone_Registration проверяет несколько пустых телефонов и конфликт непустого номера при регистрации.
func (s *UserSuite) TestPhone_Registration() {
	for i, phone := range []string{"", " \t\n", "+70000000012", "+70000000012"} {
		id := int64(i + 1)
		body := map[string]any{"email": string(rune('a'+i)) + "@example.com", "nickname": string(rune('a' + i)), "phone": phone}
		response := s.profileWrite(id, http.MethodPost, "/users/register", body)
		if i == 3 {
			s.Equal(http.StatusConflict, response.Code, response.Body.String())
			s.Contains(response.Body.String(), service.ErrPhoneConflict.Error())
			continue
		}
		s.Require().Equal(http.StatusCreated, response.Code, response.Body.String())
		user, err := s.Container.UserService.GetUserByID(s.Ctx, id)
		s.Require().NoError(err)
		if i < 2 {
			s.Nil(user.Phone)
		} else {
			s.Require().NotNil(user.Phone)
			s.Equal(phone, *user.Phone)
		}
	}
}

// TestPhone_LegacyMigration проверяет исправление старых пустых строк без изменения настоящего номера.
func (s *UserSuite) TestPhone_LegacyMigration() {
	// Все символы Unicode White_Space, пустая строка, обычный номер и значимый zero-width space.
	phonesToStore := []string{"", " \t\n", "\u0085", "\u00a0", "\u1680", "\u2000", "\u2001", "\u2002", "\u2003", "\u2004", "\u2005", "\u2006", "\u2007", "\u2008", "\u2009", "\u200a", "\u2028", "\u2029", "\u202f", "\u205f", "\u3000", "\v\f\r", "+70000000012", "\u200b"}
	for i, phone := range phonesToStore {
		id := int64(i + 1)
		s.createTestUser(id, string(rune('a'+i))+"@example.com", string(rune('a'+i)), "User")
		s.Require().NoError(s.GetDB().Exec("UPDATE users SET phone = ? WHERE id = ?", phone, id).Error)
	}
	migration, err := os.ReadFile("../internal/repository/migrations/000003_normalize_empty_phone.up.sql")
	s.Require().NoError(err)
	s.Require().NoError(s.GetDB().Exec(string(migration)).Error)
	s.Require().NoError(s.GetDB().Exec(string(migration)).Error) // Повторный запуск безопасен.
	var phones []sql.NullString
	s.Require().NoError(s.GetDB().Table("users").Order("id").Pluck("phone", &phones).Error)
	s.Require().Len(phones, len(phonesToStore))
	for i, original := range phonesToStore {
		if strings.TrimSpace(original) == "" {
			s.False(phones[i].Valid, "legacy phone %q must become NULL", original)
		} else {
			s.Equal(sql.NullString{String: original, Valid: true}, phones[i])
		}
	}
}
