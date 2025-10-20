package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Используем модели БД и мапперы для преобразования

// UserRepository реализует интерфейс для работы с пользователями в PostgreSQL через GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	dbUser := load(user)
	if err := db.GetTx(ctx, r.db).WithContext(ctx).Create(dbUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return extract(dbUser), nil
}

// CreateOrUpdate создает или обновляет пользователя
func (r *UserRepository) CreateOrUpdate(ctx context.Context, user *models.User) error {
	dbUser := load(user)
	return db.GetTx(ctx, r.db).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname_cashed", "name_cashed", "photo_uuid_cashed"}),
	}).Create(dbUser).Error
}

// BatchCreateOrUpdate создает или обновляет пользователей
func (r *UserRepository) BatchCreateOrUpdate(ctx context.Context, users []*models.User) error {
	dbUsers := make([]*User, len(users))
	for i, user := range users {
		dbUsers[i] = load(user)
	}
	return db.GetTx(ctx, r.db).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname_cashed", "name_cashed", "photo_uuid_cashed"}),
	}).Create(dbUsers).Error
}

func (r *UserRepository) BatchCreate(ctx context.Context, users []*models.User) error {
	dbUsers := make([]*User, len(users))
	for i, user := range users {
		dbUsers[i] = load(user)
	}
	return db.GetTx(ctx, r.db).WithContext(ctx).Create(dbUsers).Error
}

// GetByExternalUserIDs находит пользователей по UserID (ID из сервиса идентификации)
func (r *UserRepository) GetByExternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error) {
	var dbUsers []User
	err := r.db.WithContext(ctx).Where("user_id IN ?", ids).Find(&dbUsers).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	return extractSlice(dbUsers), nil
}

// GetByInternalUserIDs находит пользователей по UserID (ID из сервиса идентификации)
func (r *UserRepository) GetByInternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error) {
	var dbUsers []User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&dbUsers).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	return extractSlice(dbUsers), nil
}

func (r *UserRepository) GetByInternalUserID(ctx context.Context, id int64) (*models.User, error) {
	var dbUser User
	err := r.db.WithContext(ctx).First(&dbUser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	return extract(&dbUser), nil
}

// GetByExternalUserID находит пользователя по UserID (ID из сервиса идентификации)
func (r *UserRepository) GetByExternalUserID(ctx context.Context, userID int64) (*models.User, error) {
	var dbUser User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	return extract(&dbUser), nil
}

// GetByEventID находит всех пользователей, связанных с мероприятием
func (r *UserRepository) GetByEventID(ctx context.Context, eventID int64) ([]models.User, error) {
	var dbUsers []User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_event ON users.id = user_event.user_id").
		Where("user_event.event_id = ?", eventID).
		Find(&dbUsers).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей мероприятия: %w", err)
	}
	return extractSlice(dbUsers), nil
}

// GetDummiesByEventID находит всех dummy-пользователей, связанных с мероприятием
func (r *UserRepository) GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error) {
	var dbUsers []User
	err := r.db.WithContext(ctx).
		Joins("JOIN user_event ON users.user_id = user_event.user_id").
		Where("user_event.event_id = ? AND users.is_dummy = true", eventID).
		Find(&dbUsers).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении dummy-пользователей мероприятия: %w", err)
	}
	return extractSlice(dbUsers), nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	dbUser := load(user)
	if err := db.GetTx(ctx, r.db).WithContext(ctx).Save(dbUser).Error; err != nil {
		return nil, fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}
	return extract(dbUser), nil
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result := db.GetTx(ctx, r.db).WithContext(ctx).Delete(&User{}, id)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("пользователь не найден")
	}
	return nil
}

// AddUserToEvent добавляет пользователя в мероприятие
func (r *UserRepository) AddUserToEvent(ctx context.Context, userID, eventID int64) error {

	// Добавляем связь между пользователем и мероприятием
	err := db.GetTx(ctx, r.db).WithContext(ctx).
		Exec("INSERT INTO user_event (user_id, event_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			userID, eventID).Error
	if err != nil {
		return fmt.Errorf("ошибка при добавлении пользователя в мероприятие: %w", err)
	}

	return nil
}

// AddUsersToEvent добавляет пользователя в мероприятие
func (r *UserRepository) AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error {
	if len(ids) == 0 {
		return nil
	}

	var entries []UserEvent
	for _, id := range ids {
		entries = append(entries, UserEvent{UserID: id, EventID: eventID})
	}

	err := db.GetTx(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&entries).Error

	if err != nil {
		return fmt.Errorf("ошибка при добавлении пользователей в мероприятие: %w", err)
	}

	return nil
}

// RemoveUserFromEvent удаляет пользователя из мероприятия
func (r *UserRepository) RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error {
	result := db.GetTx(ctx, r.db).WithContext(ctx).
		Exec("DELETE FROM user_event WHERE user_id = ? AND event_id = ?", userID, eventID)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении пользователя из мероприятия: %w", result.Error)
	}
	return nil
}
