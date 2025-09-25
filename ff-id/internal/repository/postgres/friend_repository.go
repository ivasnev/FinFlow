package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

// FriendRepository реализует FriendRepositoryInterface
type FriendRepository struct {
	db *gorm.DB
}

// NewFriendRepository создает новый репозиторий для работы с друзьями
func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{
		db: db,
	}
}

// AddFriend добавляет пользователя в друзья (создает заявку)
func (r *FriendRepository) AddFriend(ctx context.Context, userID, friendID int64) error {
	friend := &models.UserFriend{
		UserID:   userID,
		FriendID: friendID,
		Status:   dto.FriendStatusPending,
	}

	result := r.db.WithContext(ctx).Create(friend)
	return result.Error
}

// UpdateFriendStatus обновляет статус дружбы
func (r *FriendRepository) UpdateFriendStatus(ctx context.Context, userID, friendID int64, status string) error {
	result := r.db.WithContext(ctx).
		Model(&models.UserFriend{}).
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
		result := tx.Model(&models.UserFriend{}).
			Where("user_id = ? AND friend_id = ?", friendID, userID).
			Update("status", dto.FriendStatusAccepted)

		if result.RowsAffected == 0 {
			return errors.New("friend request not found")
		}

		// Создаем обратную связь
		mutualFriend := &models.UserFriend{
			UserID:   userID,
			FriendID: friendID,
			Status:   dto.FriendStatusAccepted,
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
			Delete(&models.UserFriend{})

		// Удаляем связь в обратную сторону
		result2 := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).
			Delete(&models.UserFriend{})

		// Проверяем, что хотя бы одна запись была удалена
		if result1.RowsAffected == 0 && result2.RowsAffected == 0 {
			return errors.New("friend relationship not found")
		}

		return nil
	})
}

// GetFriendRelation получает информацию о связи дружбы между пользователями
func (r *FriendRepository) GetFriendRelation(ctx context.Context, userID, friendID int64) (*models.UserFriend, error) {
	var relation models.UserFriend

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&relation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("friend relationship not found")
		}
		return nil, result.Error
	}

	return &relation, nil
}

// GetFriends получает список друзей пользователя с пагинацией и фильтрацией
func (r *FriendRepository) GetFriends(ctx context.Context, userID int64, page, pageSize int, friendName, status string) ([]models.UserFriend, int64, error) {
	var friends []models.UserFriend
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.UserFriend{}).
		Joins("Friend"). // Всегда прогружаем информацию о друге
		Where("user_id = ?", userID)

	// Применяем фильтр по статусу, если он задан
	if status != "" {
		query = query.Where("status = ?", status)
	} else {
		// По умолчанию показываем только принятые дружеские связи
		query = query.Where("status = ?", dto.FriendStatusAccepted)
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
		Find(&friends).Error

	return friends, totalCount, err
}

// GetFriendRequests получает список заявок в друзья пользователя
func (r *FriendRepository) GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) ([]models.UserFriend, int64, error) {
	var requests []models.UserFriend
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.UserFriend{})

	if incoming {
		// Входящие заявки (где пользователь - потенциальный друг)
		query = query.Where("friend_id = ? AND status = ?", userID, dto.FriendStatusPending)
		// Прогружаем инициатора заявки (того, кто отправил)
		query = query.Joins("User")
	} else {
		// Исходящие заявки (где пользователь - инициатор)
		query = query.Where("user_id = ? AND status = ?", userID, dto.FriendStatusPending)
		// Прогружаем получателя заявки
		query = query.Joins("Friend")
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
		Find(&requests).Error

	return requests, totalCount, err
}

// GetFriendRelationWithPreload получает информацию о связи дружбы между пользователями с предзагрузкой связей
func (r *FriendRepository) GetFriendRelationWithPreload(ctx context.Context, userID, friendID int64, preloadUser, preloadFriend bool) (*models.UserFriend, error) {
	var relation models.UserFriend

	query := r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID)

	// Предзагружаем необходимые связи
	if preloadUser {
		query = query.Joins("User")
	}

	if preloadFriend {
		query = query.Joins("Friend")
	}

	result := query.First(&relation)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("friend relationship not found")
		}
		return nil, result.Error
	}

	return &relation, nil
}
