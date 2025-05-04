package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// IconServiceInterface интерфейс для работы с иконками на уровне бизнес-логики
type IconServiceInterface interface {
	GetIcons(ctx context.Context) ([]dto.IconFullDTO, error)
	GetIconByID(ctx context.Context, id uint) (*dto.IconFullDTO, error)
	CreateIcon(ctx context.Context, icon *dto.IconFullDTO) (*dto.IconFullDTO, error)
	UpdateIcon(ctx context.Context, id uint, icon *dto.IconFullDTO) (*dto.IconFullDTO, error)
	DeleteIcon(ctx context.Context, id uint) error
}

// CategoryServiceInterface интерфейс для работы с категориями на уровне бизнес-логики
type CategoryServiceInterface interface {
	GetCategories(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error)
	GetCategoryByID(ctx context.Context, id int, categoryType string) (*dto.CategoryDTO, error)
	CreateCategory(ctx context.Context, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error)
	UpdateCategory(ctx context.Context, id int, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error)
	DeleteCategory(ctx context.Context, id int, categoryType string) error
	GetCategoryTypes() ([]string, error)
}

// EventServiceInterface интерфейс для работы с мероприятиями на уровне бизнес-логики
type EventServiceInterface interface {
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id int64) (*models.Event, error)
	CreateEvent(ctx context.Context, request *dto.EventRequest) (*dto.EventResponse, error)
	UpdateEvent(ctx context.Context, id int64, request *dto.EventRequest) (*dto.EventResponse, error)
	DeleteEvent(ctx context.Context, id int64) error
}

// ActivityServiceInterface интерфейс для работы с активностями на уровне бизнес-логики
type ActivityServiceInterface interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetActivityByID(ctx context.Context, id int) (*models.Activity, error)
	CreateActivity(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	UpdateActivity(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	DeleteActivity(ctx context.Context, id int) error
}

// UserServiceInterface определяет методы для работы с пользователями
type UserServiceInterface interface {
	// CreateUser создает нового пользователя
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)

	// CreateDummyUser создает нового dummy-пользователя для мероприятия
	CreateDummyUser(ctx context.Context, name string, eventID int64) (*models.User, error)

	// BatchCreateDummyUsers создает dummy-пользователей для мероприятия
	BatchCreateDummyUsers(ctx context.Context, names []string, eventID int64) ([]*models.User, error)

	// GetUserByID получает пользователя по внутреннему ID
	GetUserByID(ctx context.Context, id int64) (*models.User, error)

	// GetUserByUserID получает пользователя по UserID (ID из сервиса идентификации)
	GetUserByUserID(ctx context.Context, userID int64) (*models.User, error)

	// GetUsersByUserIDs получает пользователей по UserID (ID из сервиса идентификации)
	GetUsersByUserIDs(ctx context.Context, userIDs []int64) ([]models.User, error)

	// GetUsersByEventID получает всех пользователей мероприятия
	GetUsersByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// GetDummiesByEventID получает всех dummy-пользователей мероприятия
	GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, id int64) error

	// AddUsersToEvent добавляет пользователей в мероприятие
	AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error

	// RemoveUserFromEvent удаляет пользователя из мероприятия
	RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error

	// SyncUserWithIDService синхронизирует данные пользователя с ID-сервисом
	SyncUserWithIDService(ctx context.Context, userID int64) (*models.User, error)

	// BatchSyncUsersWithIDService синхронизирует данные пользователей с ID-сервисом
	BatchSyncUsersWithIDService(ctx context.Context, userIDs []int64) error

	// GetNotExistsUsers получает пользователей, которые не существуют в базе данных
	GetNotExistsUsers(ctx context.Context, ids []int64) ([]int64, error)
}
