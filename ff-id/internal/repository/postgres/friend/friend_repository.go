package friend

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"gorm.io/gorm"
)

const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
)

// FriendRepository реализует интерфейс repository.Friend для работы с друзьями в PostgreSQL через GORM
type FriendRepository struct {
	db *gorm.DB
}

// NewFriendRepository создает новый репозиторий для работы с друзьями
func NewFriendRepository(db *gorm.DB) repository.Friend {
	return &FriendRepository{
		db: db,
	}
}

// AddFriend добавляет пользователя в друзья (создает заявку)
func (r *FriendRepository) AddFriend(ctx context.Context, userID, friendID int64) error {
	friend := &UserFriend{
		UserID:   userID,
		FriendID: friendID,
		Status:   FriendStatusPending,
	}

	result := r.db.WithContext(ctx).Create(friend)
	return result.Error
}

// UpdateFriendStatus обновляет статус дружбы
func (r *FriendRepository) UpdateFriendStatus(ctx context.Context, userID, friendID int64, status string) error {
	result := r.db.WithContext(ctx).
		Model(&UserFriend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("status", status)

	if result.RowsAffected == 0 {
		return errors.New("friend relationship not found")
	}

	return result.Error
}

// CreateMutualFriendship создает взаимную дружбу (при принятии заявки)
func (r *FriendRepository) CreateMutualFriendship(ctx context.Context, userID, friendID int64) error {
	// Используем транзакцию для атомарного обновления статуса и создания взаимной связи
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Обновляем статус исходной заявки
		result := tx.Model(&UserFriend{}).
			Where("user_id = ? AND friend_id = ?", friendID, userID).
			Update("status", FriendStatusAccepted)

		if result.RowsAffected == 0 {
			return errors.New("friend request not found")
		}

		// Создаем обратную связь
		mutualFriend := &UserFriend{
			UserID:   userID,
			FriendID: friendID,
			Status:   FriendStatusAccepted,
		}

		if err := tx.Create(mutualFriend).Error; err != nil {
			return err
		}

		return nil
	})
}

// RemoveFriend удаляет пользователя из друзей
func (r *FriendRepository) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	// Используем транзакцию для удаления двух записей
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удаляем связь в одну сторону
		result1 := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).
			Delete(&UserFriend{})

		// Удаляем связь в обратную сторону
		result2 := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&UserFriend{})

		// Проверяем, что хотя бы одна запись была удалена
		if result1.RowsAffected == 0 && result2.RowsAffected == 0 {
			return errors.New("friend relationship not found")
		}

		return nil
	})
}

// GetFriendRelation получает информацию о связи дружбы между пользователями
func (r *FriendRepository) GetFriendRelation(ctx context.Context, userID, friendID int64) (*models.UserFriend, error) {
	var relation UserFriend

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&relation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("friend relationship not found")
		}
		return nil, result.Error
	}

	return ExtractUserFriend(&relation), nil
}

// GetFriendRelationWithPreload получает информацию о связи дружбы между пользователями с предзагрузкой связей
func (r *FriendRepository) GetFriendRelationWithPreload(ctx context.Context, userID, friendID int64, preloadUser, preloadFriend bool) (*models.UserFriend, error) {
	var relation UserFriend
	var dbUser User
	var dbFriend User

	query := r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID)

	// Предзагружаем необходимые связи
	if preloadUser {
		query = query.Joins("JOIN users AS user_table ON user_table.id = user_friends.user_id")
	}

	if preloadFriend {
		query = query.Joins("JOIN users AS friend_table ON friend_table.id = user_friends.friend_id")
	}

	result := query.First(&relation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("friend relationship not found")
		}
		return nil, result.Error
	}

	domainFriend := ExtractUserFriend(&relation)

	// Загружаем связанных пользователей если требуется
	if preloadUser {
		if err := r.db.WithContext(ctx).First(&dbUser, relation.UserID).Error; err == nil {
			domainFriend.User = ExtractUser(&dbUser)
		}
	}

	if preloadFriend {
		if err := r.db.WithContext(ctx).First(&dbFriend, relation.FriendID).Error; err == nil {
			domainFriend.Friend = ExtractUser(&dbFriend)
		}
	}

	return domainFriend, nil
}

// GetFriends получает список друзей пользователя с пагинацией и фильтрацией
func (r *FriendRepository) GetFriends(ctx context.Context, userID int64, page, pageSize int, friendName, status string) ([]models.UserFriend, int64, error) {
	var dbFriends []UserFriend
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&UserFriend{}).
		Where("user_friends.user_id = ?", userID)

	// Применяем фильтр по статусу, если он задан
	if status != "" {
		query = query.Where("user_friends.status = ?", status)
	} else {
		// По умолчанию показываем только принятые дружеские связи
		query = query.Where("user_friends.status = ?", FriendStatusAccepted)
	}

	// Применяем фильтр по имени, если он задан
	if friendName != "" {
		query = query.Joins("JOIN users ON users.id = user_friends.friend_id").
			Where("LOWER(users.name) ILIKE LOWER(?)", fmt.Sprintf("%%%s%%", friendName))
	}

	// Получаем общее количество записей
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Получаем записи с пагинацией
	offset := (page - 1) * pageSize
	err := query.Order("user_friends.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&dbFriends).Error

	if err != nil {
		return nil, 0, err
	}

	// Загружаем информацию о друзьях
	friends := make([]models.UserFriend, len(dbFriends))
	for i, dbFriend := range dbFriends {
		friends[i] = *ExtractUserFriend(&dbFriend)

		// Загружаем информацию о друге
		var dbUser User
		if err := r.db.WithContext(ctx).First(&dbUser, dbFriend.FriendID).Error; err == nil {
			friends[i].Friend = ExtractUser(&dbUser)
		}
	}

	return friends, totalCount, nil
}

// GetFriendRequests получает список заявок в друзья пользователя
func (r *FriendRepository) GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) ([]models.UserFriend, int64, error) {
	var dbRequests []UserFriend
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&UserFriend{})

	if incoming {
		// Входящие заявки (где пользователь - потенциальный друг)
		query = query.Where("friend_id = ? AND status = ?", userID, FriendStatusPending)
	} else {
		// Исходящие заявки (где пользователь - инициатор)
		query = query.Where("user_id = ? AND status = ?", userID, FriendStatusPending)
	}

	// Получаем общее количество записей
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Получаем записи с пагинацией
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&dbRequests).Error

	if err != nil {
		return nil, 0, err
	}

	// Преобразуем в доменные модели и загружаем связи
	requests := make([]models.UserFriend, len(dbRequests))
	for i, dbRequest := range dbRequests {
		requests[i] = *ExtractUserFriend(&dbRequest)

		if incoming {
			// Для входящих заявок загружаем инициатора (User)
			var dbUser User
			if err := r.db.WithContext(ctx).First(&dbUser, dbRequest.UserID).Error; err == nil {
				requests[i].User = ExtractUser(&dbUser)
			}
		} else {
			// Для исходящих заявок загружаем получателя (Friend)
			var dbFriend User
			if err := r.db.WithContext(ctx).First(&dbFriend, dbRequest.FriendID).Error; err == nil {
				requests[i].Friend = ExtractUser(&dbFriend)
			}
		}
	}

	return requests, totalCount, nil
}
