package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	customErrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
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
	err := db.GetTx(ctx, r.db).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nickname_cashed", "name_cashed", "photo_uuid_cashed"}),
	}).Create(dbUser).Error
	if err != nil {
		return fmt.Errorf("ошибка при создании или обновлении пользователя: %w", err)
	}
	user.ID = dbUser.ID
	return nil
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

	err := db.GetTx(ctx, r.db).WithContext(ctx).Create(dbUsers).Error
	if err != nil {
		return err
	}

	// Заполняем ID в исходных моделях после создания
	for i, dbUser := range dbUsers {
		users[i].ID = dbUser.ID
	}

	return nil
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
			return nil, customErrors.NewEntityNotFoundError(strconv.FormatInt(id, 10), "user")
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
			return nil, repository.ErrUserNotFound
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
		Joins("JOIN user_event ON users.id = user_event.user_id").
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
		return customErrors.NewEntityNotFoundError(strconv.FormatInt(id, 10), "user")
	}
	return nil
}

// changeMembership объединяет фактическое изменение участия и журнал в общей транзакции.
func (r *UserRepository) changeMembership(ctx context.Context, ids []int64, eventID int64, added bool) error {
	return db.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := db.GetTx(ctx, r.db).WithContext(ctx)
		if err := history.LockEvent(tx, eventID); err != nil {
			return err
		}
		for _, id := range ids {
			var result *gorm.DB
			if added {
				result = tx.Exec("INSERT INTO user_event(user_id,event_id) VALUES(?,?) ON CONFLICT DO NOTHING", id, eventID)
			} else {
				result = tx.Exec("DELETE FROM user_event WHERE user_id=? AND event_id=?", id, eventID)
			}
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected > 0 {
				if !added {
					var links []struct{ ID int64 }
					if err := tx.Raw("UPDATE event_invitations SET revoked_at=CURRENT_TIMESTAMP WHERE event_id=? AND created_by=(SELECT user_id FROM users WHERE id=? AND NOT is_dummy) AND revoked_at IS NULL RETURNING id", eventID, id).Scan(&links).Error; err != nil {
						return err
					}
					for _, link := range links {
						if err := history.Record(ctx, tx, eventID, "invitation.revoked", link.ID, map[string]string{"Приглашение": "Действует"}, map[string]string{"Приглашение": "Создатель больше не участвует в событии"}); err != nil {
							return err
						}
					}
				}
				if err := history.Member(ctx, tx, eventID, id, added); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// AddUserToEvent сохраняет добавление одного участника.
func (r *UserRepository) AddUserToEvent(ctx context.Context, userID, eventID int64) error {
	return r.changeMembership(ctx, []int64{userID}, eventID, true)
}

// AddUsersToEvent сохраняет добавление участников без дублирования истории.
func (r *UserRepository) AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error {
	return r.changeMembership(ctx, ids, eventID, true)
}

// RemoveUserFromEvent сохраняет удаление участника вместе с его именем в истории.
func (r *UserRepository) RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error {
	return r.changeMembership(ctx, []int64{userID}, eventID, false)
}

// InTransaction объединяет создание гостя, его участие и журнал в одном commit.
func (r *UserRepository) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return db.WithTx(ctx, r.db, fn)
}
