package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// CategoryRepository интерфейс для работы с категориями
type CategoryRepository interface {
	GetAll(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error)
	GetByID(ctx context.Context, categoryType string, id int) (*dto.CategoryDTO, error)
	Create(ctx context.Context, categoryType string, category *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Update(ctx context.Context, categoryType string, category *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Delete(ctx context.Context, categoryType string, id int) error
	GetCategoryTypes() ([]string, error)
}

// EventRepository интерфейс для работы с мероприятиями
type EventRepository interface {
	GetAll(ctx context.Context) ([]models.Event, error)
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	Create(ctx context.Context, event *models.Event) error
	Update(ctx context.Context, id int64, event *models.Event) error
	Delete(ctx context.Context, id int64) error
}

// ActivityRepository интерфейс для работы с активностями
type ActivityRepository interface {
	GetByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetByID(ctx context.Context, id int) (*models.Activity, error)
	Create(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	Update(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	Delete(ctx context.Context, id int) error
}

// UserRepositoryInterface определяет методы для работы с пользователями
type UserRepositoryInterface interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) (*models.User, error)

	// BatchCreate создает пользователей
	BatchCreate(ctx context.Context, users []*models.User) error

	// BatchCreateOrUpdate создает или обновляет пользователей
	BatchCreateOrUpdate(ctx context.Context, users []*models.User) error

	// CreateOrUpdate создает или обновляет пользователя
	CreateOrUpdate(ctx context.Context, user *models.User) error

	// GetByExternalUserIDs находит пользователей по UserID (ID из сервиса идентификации)
	GetByExternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error)

	// GetByInternalUserIDs находит пользователей по UserID (ID из сервиса идентификации)
	GetByInternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error)

	// GetByInternalUserID находит пользователя по внутреннему ID
	GetByInternalUserID(ctx context.Context, id int64) (*models.User, error)

	// GetByExternalUserID находит пользователя по UserID (ID из сервиса идентификации)
	GetByExternalUserID(ctx context.Context, userID int64) (*models.User, error)

	// GetByEventID находит всех пользователей, связанных с мероприятием
	GetByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// GetDummiesByEventID находит всех dummy-пользователей, связанных с мероприятием
	GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) (*models.User, error)

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error

	// AddUserToEvent добавляет пользователя в мероприятие
	AddUserToEvent(ctx context.Context, userID, eventID int64) error

	// AddUsersToEvent добавляет пользователя в мероприятие
	AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error

	// RemoveUserFromEvent удаляет пользователя из мероприятия
	RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error
}

// TransactionRepository интерфейс для работы с транзакциями
type TransactionRepository interface {
	// Получение транзакций
	GetTransactionsByEventID(eventID int64) ([]models.Transaction, error)
	GetTransactionByID(id int) (*models.Transaction, error)

	// Управление транзакциями
	CreateTransaction(tx *models.Transaction) error
	UpdateTransaction(tx *models.Transaction) error
	DeleteTransaction(id int) error

	// Работа с долями транзакций
	GetSharesByTransactionID(transactionID int) ([]models.TransactionShare, error)
	CreateTransactionShares(shares []models.TransactionShare) error
	DeleteSharesByTransactionID(transactionID int) error

	// Работа с долгами
	GetDebtsByTransactionID(transactionID int) ([]models.Debt, error)
	GetDebtsByEventID(eventID int64) ([]models.Debt, error)
	GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]models.Debt, error)
	GetDebtsByEventIDToUser(eventID int64, userID int64) ([]models.Debt, error)
	CreateDebts(debts []models.Debt) error
	DeleteDebtsByTransactionID(transactionID int) error

	// Работа с оптимизированными долгами
	GetOptimizedDebtsByEventID(eventID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByEventIDWithUsers(eventID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByUserID(eventID, userID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByUserIDWithUsers(eventID, userID int64) ([]models.OptimizedDebt, error)
	SaveOptimizedDebts(eventID int64, debts []models.OptimizedDebt) error
	DeleteOptimizedDebtsByEventID(eventID int64) error
}
