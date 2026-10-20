package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service/password"
)

// TestConfiguredPasswordMinimum проверяет настройку сервиса смены пароля до обращения к репозиторию.
func TestConfiguredPasswordMinimum(t *testing.T) {
	controller := gomock.NewController(t)
	repo := mock.NewMockUser(controller)
	svc := NewUserService(repo, 12, nil)
	value := "12345678901"
	if _, err := svc.UpdateUser(context.Background(), 1, service.UserUpdateData{Password: &value}); !errors.Is(err, password.ErrInvalid) {
		t.Fatalf("configured minimum ignored: %v", err)
	}
}
