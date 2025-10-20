package friend

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// Константы для пагинации
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	DefaultPage     = 1
)

// FriendService реализует FriendServiceInterface
type FriendService struct {
	friendRepo repository.Friend
	userRepo   repository.User
}

// NewFriendService создает новый FriendService
func NewFriendService(
	friendRepo repository.Friend,
	userRepo repository.User,
) service.FriendServiceInterface {
	return &FriendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
	}
}

// AddFriend создает заявку на добавление в друзья
func (s *FriendService) AddFriend(ctx context.Context, userID int64, req service.AddFriendRequest) error {
	User, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("пользователь не найден")
	}
	// Проверяем, что пользователь не пытается добавить сам себя
	if User.Nickname == req.FriendNickname {
		return errors.New("нельзя добавить самого себя в друзья")
	}

	// Проверяем, что пользователь с friendID существует
	friendUser, err := s.userRepo.GetByNickname(ctx, req.FriendNickname)
	if err != nil {
		return errors.New("пользователь для добавления в друзья не найден")
	}

	// Проверяем, существует ли уже отношение дружбы с предзагрузкой связи Friend
	relation, err := s.friendRepo.GetFriendRelationWithPreload(ctx, userID, friendUser.ID, false, true)
	if err == nil {
		// Получаем имя друга для улучшения сообщений об ошибках
		friendName := friendUser.Nickname
		if friendUser.Name.Valid {
			friendName = friendUser.Name.String
		}

		// Отношение уже существует
		if relation.Status == service.FriendStatusAccepted {
			return fmt.Errorf("пользователь %s уже в списке ваших друзей", friendName)
		} else if relation.Status == service.FriendStatusPending {
			return fmt.Errorf("заявка в друзья пользователю %s уже отправлена", friendName)
		} else if relation.Status == service.FriendStatusBlocked {
			return fmt.Errorf("пользователь %s заблокирован", friendName)
		}
	}

	// Проверяем, не заблокировал ли вас этот пользователь
	blockedRelation, err := s.friendRepo.GetFriendRelationWithPreload(ctx, friendUser.ID, userID, false, false)
	if err == nil && blockedRelation.Status == service.FriendStatusBlocked {
		return errors.New("невозможно отправить заявку этому пользователю")
	}

	// Добавляем заявку в друзья
	return s.friendRepo.AddFriend(ctx, userID, friendUser.ID)
}

// AcceptFriendRequest принимает заявку в друзья
func (s *FriendService) AcceptFriendRequest(ctx context.Context, userID int64, req service.FriendActionRequest) error {
	// Проверяем, что заявка существует, и предзагружаем информацию об отправителе
	relation, err := s.friendRepo.GetFriendRelationWithPreload(ctx, req.UserID, userID, false, false)
	if err != nil {
		return errors.New("заявка в друзья не найдена")
	}

	if relation.Status != service.FriendStatusPending {
		return errors.New("некорректный статус заявки")
	}

	// Создаем взаимную дружбу
	return s.friendRepo.CreateMutualFriendship(ctx, userID, req.UserID)
}

// RejectFriendRequest отклоняет заявку в друзья
func (s *FriendService) RejectFriendRequest(ctx context.Context, userID int64, req service.FriendActionRequest) error {
	// Проверяем, что заявка существует, и предзагружаем информацию об отправителе
	relation, err := s.friendRepo.GetFriendRelationWithPreload(ctx, req.UserID, userID, false, false)
	if err != nil {
		return errors.New("заявка в друзья не найдена")
	}

	if relation.Status != service.FriendStatusPending {
		return errors.New("некорректный статус заявки")
	}

	// Обновляем статус на "отклонено"
	return s.friendRepo.UpdateFriendStatus(ctx, req.UserID, userID, service.FriendStatusRejected)
}

// BlockUser блокирует пользователя
func (s *FriendService) BlockUser(ctx context.Context, userID int64, req service.FriendActionRequest) error {
	// Проверяем, что пользователь существует
	friendUser, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return errors.New("пользователь не найден")
	}

	// Проверяем текущее отношение с предзагрузкой
	_, err = s.friendRepo.GetFriendRelationWithPreload(ctx, userID, req.UserID, false, false)

	if err == nil {
		// Отношение существует, обновляем статус на "заблокирован"
		return s.friendRepo.UpdateFriendStatus(ctx, userID, req.UserID, service.FriendStatusBlocked)
	}

	// Отношение не существует, создаем новое с статусом "заблокирован"
	blockRelation := &service.AddFriendRequest{
		FriendNickname: friendUser.Nickname,
	}
	if err := s.AddFriend(ctx, userID, *blockRelation); err != nil {
		return err
	}

	return s.friendRepo.UpdateFriendStatus(ctx, userID, req.UserID, service.FriendStatusBlocked)
}

// RemoveFriend удаляет пользователя из друзей
func (s *FriendService) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	// Удаляем дружбу
	return s.friendRepo.RemoveFriend(ctx, userID, friendID)
}

// GetFriendStatus получает статус дружбы между пользователями
func (s *FriendService) GetFriendStatus(ctx context.Context, userID, friendID int64) (string, error) {
	relation, err := s.friendRepo.GetFriendRelationWithPreload(ctx, userID, friendID, false, false)
	if err != nil {
		return "", err
	}

	return relation.Status, nil
}

// GetFriends получает список друзей пользователя
func (s *FriendService) GetFriends(ctx context.Context, nickname string, params service.FriendsQueryParams) (*service.FriendsListResponse, error) {
	// Находим пользователя по никнейму
	user, err := s.userRepo.GetByNickname(ctx, nickname)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Применяем значения по умолчанию для пагинации
	page := params.Page
	if page < 1 {
		page = DefaultPage
	}

	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Получаем друзей с пагинацией
	friends, total, err := s.friendRepo.GetFriends(ctx, user.ID, page, pageSize, params.FriendName, params.Status)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &service.FriendsListResponse{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Objects:  make([]service.FriendDTO, 0, len(friends)),
	}

	// Преобразуем friends в FriendDTO
	for _, friend := range friends {
		var photoID uuid.UUID
		if friend.Friend.AvatarID.Valid {
			photoID = friend.Friend.AvatarID.UUID
		}

		var name string
		if friend.Friend.Name.Valid {
			name = friend.Friend.Name.String
		} else {
			name = friend.Friend.Nickname
		}

		friendDTO := service.FriendDTO{
			UserID:  friend.FriendID,
			PhotoID: photoID,
			Name:    name,
			Status:  friend.Status,
		}
		response.Objects = append(response.Objects, friendDTO)
	}

	return response, nil
}

// GetFriendRequests получает список заявок в друзья
func (s *FriendService) GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) (*service.FriendsListResponse, error) {
	// Применяем значения по умолчанию для пагинации
	if page < 1 {
		page = DefaultPage
	}

	if pageSize < 1 {
		pageSize = DefaultPageSize
	} else if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Получаем заявки в друзья
	requests, total, err := s.friendRepo.GetFriendRequests(ctx, userID, page, pageSize, incoming)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := &service.FriendsListResponse{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Objects:  make([]service.FriendDTO, 0, len(requests)),
	}

	// Преобразуем requests в FriendDTO
	for _, request := range requests {
		var userModel *models.User

		if incoming {
			// Для входящих заявок используем связь User - того, кто отправил заявку
			userModel = &request.User
		} else {
			// Для исходящих заявок используем связь Friend - того, кому отправлена заявка
			userModel = &request.Friend
		}

		var photoID uuid.UUID
		if userModel.AvatarID.Valid {
			photoID = userModel.AvatarID.UUID
		}

		var name string
		if userModel.Name.Valid {
			name = userModel.Name.String
		} else {
			name = userModel.Nickname
		}

		// Определяем ID пользователя для отображения
		var displayUserID int64
		if incoming {
			displayUserID = request.UserID // ID того, кто отправил заявку
		} else {
			displayUserID = request.FriendID // ID того, кому отправлена заявка
		}

		friendDTO := service.FriendDTO{
			UserID:  displayUserID,
			PhotoID: photoID,
			Name:    name,
			Status:  request.Status,
		}
		response.Objects = append(response.Objects, friendDTO)
	}

	return response, nil
}
