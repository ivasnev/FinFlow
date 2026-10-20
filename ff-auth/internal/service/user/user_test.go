package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	serviceMock "github.com/ivasnev/FinFlow/ff-auth/internal/service/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestUserService_CurrentIdentity проверяет получение актуальных полей вместо локальной копии.
func TestUserService_CurrentIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockUser(ctrl)
	identity := serviceMock.NewMockIdentity(ctrl)
	svc := NewUserService(repo, 8, identity)
	ctx := context.Background()
	for _, byNickname := range []bool{false, true} {
		if byNickname {
			identity.EXPECT().FindIdentity(ctx, "new", ffid.LookupNickname).Return(&ffid.UserDTO{ID: 1}, nil)
		}
		repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1, Email: "old@example.com", Nickname: "old"}, nil)
		identity.EXPECT().GetIdentity(ctx, int64(1)).Return(&ffid.UserDTO{ID: 1, Email: "new@example.com", Nickname: "new", UpdatedAt: 1234567890}, nil)
		repo.EXPECT().GetRoles(ctx, int64(1)).Return([]models.RoleEntity{{Name: "user"}}, nil)
		var got *service.UserData
		var err error
		if byNickname {
			got, err = svc.GetUserByNickname(ctx, "new")
		} else {
			got, err = svc.GetUserByID(ctx, 1)
		}
		require.NoError(t, err)
		require.Equal(t, "new@example.com", got.Email)
		require.Equal(t, "new", got.Nickname)
		require.Equal(t, int64(1234567890), got.UpdatedAt.Unix())
	}
	identity.EXPECT().FindIdentity(ctx, "old", ffid.LookupNickname).Return(nil, ffid.ErrNotFound)
	_, err := svc.GetUserByNickname(ctx, "old")
	require.ErrorIs(t, err, ffid.ErrNotFound)
	repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1}, nil)
	identity.EXPECT().GetIdentity(ctx, int64(1)).Return(nil, ffid.ErrUnavailable)
	_, err = svc.GetUserByID(ctx, 1)
	require.ErrorIs(t, err, ffid.ErrUnavailable)
}

// TestUserService_UpdateIdentity проверяет единственную запись email и nickname в ff-id.
func TestUserService_UpdateIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockUser(ctrl)
	identity := serviceMock.NewMockIdentity(ctrl)
	svc := NewUserService(repo, 8, identity)
	ctx := context.Background()
	email, nickname := "new@example.com", "new"
	repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1, Email: "old@example.com", Nickname: "old"}, nil)
	identity.EXPECT().UpdateIdentity(ctx, int64(1), &email, &nickname).Return(&ffid.UserDTO{ID: 1, Email: email, Nickname: nickname}, nil)
	repo.EXPECT().GetRoles(ctx, int64(1)).Return([]models.RoleEntity{{Name: "user"}}, nil)
	got, err := svc.UpdateUser(ctx, 1, service.UserUpdateData{Email: &email, Nickname: &nickname})
	require.NoError(t, err)
	require.Equal(t, email, got.Email)
	require.Equal(t, nickname, got.Nickname)
}

// TestUserService_PasswordAndIdentityFailure проверяет частичные отказы смешанного запроса и повторяемость смены пароля.
func TestUserService_PasswordAndIdentityFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockUser(ctrl)
	identity := serviceMock.NewMockIdentity(ctrl)
	svc := NewUserService(repo, 8, identity)
	ctx := context.Background()
	email, value := "new@example.com", "password123"
	repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1}, nil)
	identity.EXPECT().UpdateIdentity(ctx, int64(1), &email, nil).Return(nil, ffid.ErrUnavailable)
	_, err := svc.UpdateUser(ctx, 1, service.UserUpdateData{Email: &email, Password: &value})
	require.ErrorIs(t, err, ffid.ErrUnavailable)
	// Сценарий: ошибка локального сохранения пароля не создаёт вторую версию идентификаторов.
	repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1}, nil)
	identity.EXPECT().UpdateIdentity(ctx, int64(1), &email, nil).Return(&ffid.UserDTO{ID: 1, Email: email}, nil)
	failure := errors.New("database failure")
	repo.EXPECT().Update(ctx, gomock.Any()).Return(failure)
	_, err = svc.UpdateUser(ctx, 1, service.UserUpdateData{Email: &email, Password: &value})
	require.ErrorIs(t, err, failure)
	// Сценарий: повторная смена пароля использует актуальный профиль и сохраняет рабочий хеш.
	repo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1}, nil)
	identity.EXPECT().GetIdentity(ctx, int64(1)).Return(&ffid.UserDTO{ID: 1, Email: email}, nil)
	repo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, user *models.User) error {
		return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(value))
	})
	repo.EXPECT().GetRoles(ctx, int64(1)).Return(nil, nil)
	beforePassword := time.Now().Add(-time.Second)
	got, err := svc.UpdateUser(ctx, 1, service.UserUpdateData{Password: &value})
	require.NoError(t, err)
	require.Equal(t, email, got.Email)
	require.True(t, got.UpdatedAt.After(beforePassword))
}

// TestUserService_MissingAndDelete проверяет отсутствие профиля и удаление локальных учётных данных.
func TestUserService_MissingAndDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockUser(ctrl)
	identity := serviceMock.NewMockIdentity(ctrl)
	svc := NewUserService(repo, 8, identity)
	ctx := context.Background()
	failure := errors.New("missing")
	repo.EXPECT().GetByID(ctx, int64(1)).Return(nil, failure)
	_, err := svc.UpdateUser(ctx, 1, service.UserUpdateData{})
	require.ErrorIs(t, err, failure)
	repo.EXPECT().Delete(ctx, int64(1)).Return(nil)
	require.NoError(t, svc.DeleteUser(ctx, 1))
}
