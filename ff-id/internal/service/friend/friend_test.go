package friend

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

func TestFriendService_AddFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное добавление друга", func(t *testing.T) {
		user := &models.User{
			ID:       userID,
			Nickname: "user1",
		}

		friendUser := &models.User{
			ID:       2,
			Nickname: "user2",
		}

		req := service.AddFriendRequest{
			FriendNickname: "user2",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, req.FriendNickname).
			Return(friendUser, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, userID, friendUser.ID, false, true).
			Return(nil, errors.New("not found")).
			Times(1)

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendUser.ID, userID, false, false).
			Return(nil, errors.New("not found")).
			Times(1)

		mockFriendRepo.EXPECT().
			AddFriend(ctx, userID, friendUser.ID).
			Return(nil).
			Times(1)

		err := friendService.AddFriend(ctx, userID, req)

		assert.NoError(t, err)
	})

	t.Run("нельзя добавить самого себя", func(t *testing.T) {
		user := &models.User{
			ID:       userID,
			Nickname: "user1",
		}

		req := service.AddFriendRequest{
			FriendNickname: "user1",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		err := friendService.AddFriend(ctx, userID, req)

		assert.Error(t, err)
		assert.Equal(t, "нельзя добавить самого себя в друзья", err.Error())
	})

	t.Run("пользователь для добавления не найден", func(t *testing.T) {
		user := &models.User{
			ID:       userID,
			Nickname: "user1",
		}

		req := service.AddFriendRequest{
			FriendNickname: "nonexistent",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, req.FriendNickname).
			Return(nil, errors.New("not found")).
			Times(1)

		err := friendService.AddFriend(ctx, userID, req)

		assert.Error(t, err)
		assert.Equal(t, "пользователь для добавления в друзья не найден", err.Error())
	})

	t.Run("пользователь уже в друзьях", func(t *testing.T) {
		user := &models.User{
			ID:       userID,
			Nickname: "user1",
		}

		friendUser := &models.User{
			ID:       2,
			Nickname: "user2",
			Name:     sql.NullString{String: "User Two", Valid: true},
		}

		req := service.AddFriendRequest{
			FriendNickname: "user2",
		}

		relation := &models.UserFriend{
			ID:       1,
			UserID:   userID,
			FriendID: friendUser.ID,
			Status:   service.FriendStatusAccepted,
			Friend:   *friendUser,
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, req.FriendNickname).
			Return(friendUser, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, userID, friendUser.ID, false, true).
			Return(relation, nil).
			Times(1)

		err := friendService.AddFriend(ctx, userID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "уже в списке ваших друзей")
	})
}

func TestFriendService_AcceptFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)
	friendID := int64(2)

	t.Run("успешное принятие заявки", func(t *testing.T) {
		req := service.FriendActionRequest{
			UserID: friendID,
		}

		relation := &models.UserFriend{
			ID:       1,
			UserID:   friendID,
			FriendID: userID,
			Status:   service.FriendStatusPending,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendID, userID, false, false).
			Return(relation, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			CreateMutualFriendship(ctx, userID, friendID).
			Return(nil).
			Times(1)

		err := friendService.AcceptFriendRequest(ctx, userID, req)

		assert.NoError(t, err)
	})

	t.Run("заявка не найдена", func(t *testing.T) {
		req := service.FriendActionRequest{
			UserID: friendID,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendID, userID, false, false).
			Return(nil, errors.New("not found")).
			Times(1)

		err := friendService.AcceptFriendRequest(ctx, userID, req)

		assert.Error(t, err)
		assert.Equal(t, "заявка в друзья не найдена", err.Error())
	})

	t.Run("некорректный статус заявки", func(t *testing.T) {
		req := service.FriendActionRequest{
			UserID: friendID,
		}

		relation := &models.UserFriend{
			ID:       1,
			UserID:   friendID,
			FriendID: userID,
			Status:   service.FriendStatusAccepted,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendID, userID, false, false).
			Return(relation, nil).
			Times(1)

		err := friendService.AcceptFriendRequest(ctx, userID, req)

		assert.Error(t, err)
		assert.Equal(t, "некорректный статус заявки", err.Error())
	})
}

func TestFriendService_RejectFriendRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)
	friendID := int64(2)

	t.Run("успешное отклонение заявки", func(t *testing.T) {
		req := service.FriendActionRequest{
			UserID: friendID,
		}

		relation := &models.UserFriend{
			ID:       1,
			UserID:   friendID,
			FriendID: userID,
			Status:   service.FriendStatusPending,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendID, userID, false, false).
			Return(relation, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			UpdateFriendStatus(ctx, friendID, userID, service.FriendStatusRejected).
			Return(nil).
			Times(1)

		err := friendService.RejectFriendRequest(ctx, userID, req)

		assert.NoError(t, err)
	})

	t.Run("заявка не найдена", func(t *testing.T) {
		req := service.FriendActionRequest{
			UserID: friendID,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, friendID, userID, false, false).
			Return(nil, errors.New("not found")).
			Times(1)

		err := friendService.RejectFriendRequest(ctx, userID, req)

		assert.Error(t, err)
		assert.Equal(t, "заявка в друзья не найдена", err.Error())
	})
}

func TestFriendService_RemoveFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)
	friendID := int64(2)

	t.Run("успешное удаление друга", func(t *testing.T) {
		mockFriendRepo.EXPECT().
			RemoveFriend(ctx, userID, friendID).
			Return(nil).
			Times(1)

		err := friendService.RemoveFriend(ctx, userID, friendID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления друга", func(t *testing.T) {
		expectedErr := errors.New("remove error")
		mockFriendRepo.EXPECT().
			RemoveFriend(ctx, userID, friendID).
			Return(expectedErr).
			Times(1)

		err := friendService.RemoveFriend(ctx, userID, friendID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestFriendService_GetFriendStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)
	friendID := int64(2)

	t.Run("успешное получение статуса", func(t *testing.T) {
		relation := &models.UserFriend{
			ID:       1,
			UserID:   userID,
			FriendID: friendID,
			Status:   service.FriendStatusAccepted,
		}

		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, userID, friendID, false, false).
			Return(relation, nil).
			Times(1)

		status, err := friendService.GetFriendStatus(ctx, userID, friendID)

		assert.NoError(t, err)
		assert.Equal(t, service.FriendStatusAccepted, status)
	})

	t.Run("отношение не найдено", func(t *testing.T) {
		mockFriendRepo.EXPECT().
			GetFriendRelationWithPreload(ctx, userID, friendID, false, false).
			Return(nil, errors.New("not found")).
			Times(1)

		status, err := friendService.GetFriendStatus(ctx, userID, friendID)

		assert.Error(t, err)
		assert.Empty(t, status)
	})
}

func TestFriendService_GetFriends(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	nickname := "testuser"

	t.Run("успешное получение списка друзей", func(t *testing.T) {
		user := &models.User{
			ID:       1,
			Nickname: nickname,
		}

		friendUser := models.User{
			ID:       2,
			Nickname: "friend1",
			Name:     sql.NullString{String: "Friend One", Valid: true},
			AvatarID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		}

		friends := []models.UserFriend{
			{
				ID:       1,
				UserID:   1,
				FriendID: 2,
				Status:   service.FriendStatusAccepted,
				Friend:   friendUser,
			},
		}

		params := service.FriendsQueryParams{
			Page:     1,
			PageSize: 20,
		}

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(user, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			GetFriends(ctx, user.ID, 1, 20, "", "").
			Return(friends, int64(1), nil).
			Times(1)

		result, err := friendService.GetFriends(ctx, nickname, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, 1, len(result.Objects))
		assert.Equal(t, int64(2), result.Objects[0].UserID)
		assert.Equal(t, "Friend One", result.Objects[0].Name)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		params := service.FriendsQueryParams{
			Page:     1,
			PageSize: 20,
		}

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(nil, errors.New("not found")).
			Times(1)

		result, err := friendService.GetFriends(ctx, nickname, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь не найден", err.Error())
	})

	t.Run("применение значений по умолчанию для пагинации", func(t *testing.T) {
		user := &models.User{
			ID:       1,
			Nickname: nickname,
		}

		params := service.FriendsQueryParams{
			Page:     0,
			PageSize: 0,
		}

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(user, nil).
			Times(1)

		mockFriendRepo.EXPECT().
			GetFriends(ctx, user.ID, DefaultPage, DefaultPageSize, "", "").
			Return([]models.UserFriend{}, int64(0), nil).
			Times(1)

		result, err := friendService.GetFriends(ctx, nickname, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, DefaultPage, result.Page)
		assert.Equal(t, DefaultPageSize, result.PageSize)
	})
}

func TestFriendService_GetFriendRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFriendRepo := mock.NewMockFriend(ctrl)
	mockUserRepo := mock.NewMockUser(ctrl)
	friendService := NewFriendService(mockFriendRepo, mockUserRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное получение входящих заявок", func(t *testing.T) {
		requestUser := models.User{
			ID:       2,
			Nickname: "requester",
			Name:     sql.NullString{String: "Requester", Valid: true},
			AvatarID: uuid.NullUUID{UUID: uuid.New(), Valid: true},
		}

		requests := []models.UserFriend{
			{
				ID:       1,
				UserID:   2,
				FriendID: 1,
				Status:   service.FriendStatusPending,
				User:     requestUser,
			},
		}

		mockFriendRepo.EXPECT().
			GetFriendRequests(ctx, userID, 1, 20, true).
			Return(requests, int64(1), nil).
			Times(1)

		result, err := friendService.GetFriendRequests(ctx, userID, 1, 20, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Objects))
		assert.Equal(t, int64(2), result.Objects[0].UserID)
		assert.Equal(t, "Requester", result.Objects[0].Name)
	})

	t.Run("успешное получение исходящих заявок", func(t *testing.T) {
		friendUser := models.User{
			ID:       3,
			Nickname: "friend",
			Name:     sql.NullString{String: "Friend", Valid: true},
		}

		requests := []models.UserFriend{
			{
				ID:       2,
				UserID:   1,
				FriendID: 3,
				Status:   service.FriendStatusPending,
				Friend:   friendUser,
			},
		}

		mockFriendRepo.EXPECT().
			GetFriendRequests(ctx, userID, 1, 20, false).
			Return(requests, int64(1), nil).
			Times(1)

		result, err := friendService.GetFriendRequests(ctx, userID, 1, 20, false)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Objects))
		assert.Equal(t, int64(3), result.Objects[0].UserID)
		assert.Equal(t, "Friend", result.Objects[0].Name)
	})

	t.Run("применение значений по умолчанию", func(t *testing.T) {
		mockFriendRepo.EXPECT().
			GetFriendRequests(ctx, userID, DefaultPage, DefaultPageSize, true).
			Return([]models.UserFriend{}, int64(0), nil).
			Times(1)

		result, err := friendService.GetFriendRequests(ctx, userID, 0, 0, true)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, DefaultPage, result.Page)
		assert.Equal(t, DefaultPageSize, result.PageSize)
	})
}

