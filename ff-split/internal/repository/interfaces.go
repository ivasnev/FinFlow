package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Repository представляет основной интерфейс репозитория
type Repository interface {
	CategoryRepository() CategoryRepository
	EventRepository() EventRepository
	ActivityRepository() ActivityRepository
	TransactionRepository() TransactionRepository
	TransactionTypeRepository() TransactionTypeRepository
	IconRepository() IconRepository
}

// CategoryRepository представляет интерфейс для работы с категориями
type CategoryRepository interface {
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id int) (models.Category, error)
	CreateCategory(ctx context.Context, category models.Category) (models.Category, error)
	UpdateCategory(ctx context.Context, category models.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

// EventRepository представляет интерфейс для работы с мероприятиями
type EventRepository interface {
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id int64) (models.Event, error)
	CreateEvent(ctx context.Context, event models.Event) (models.Event, error)
	UpdateEvent(ctx context.Context, event models.Event) error
	DeleteEvent(ctx context.Context, id int64) error

	// Методы для работы с участниками мероприятия
	AddUserToEvent(ctx context.Context, userEvent models.UserEvent) error
	RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error
	GetEventUsers(ctx context.Context, eventID int64) ([]models.User, error)

	// Создание искусственных пользователей
	CreateDummyUser(ctx context.Context, name string) (models.User, error)
}

// ActivityRepository представляет интерфейс для работы с событиями в мероприятии
type ActivityRepository interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetActivityByID(ctx context.Context, id int) (models.Activity, error)
	CreateActivity(ctx context.Context, activity models.Activity) (models.Activity, error)
	UpdateActivity(ctx context.Context, activity models.Activity) error
	DeleteActivity(ctx context.Context, id int) error
}

// TransactionRepository представляет интерфейс для работы с транзакциями
type TransactionRepository interface {
	GetTransactionsByEventID(ctx context.Context, eventID int64) ([]models.Transaction, error)
	GetTransactionByID(ctx context.Context, id int) (models.Transaction, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction) (models.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction models.Transaction) error
	DeleteTransaction(ctx context.Context, id int) error

	// Методы для работы с участниками транзакции
	AddUserToTransaction(ctx context.Context, userTransaction models.UserTransaction) error
	RemoveUserFromTransaction(ctx context.Context, userID int64, transactionID int) error
	GetTransactionUsers(ctx context.Context, transactionID int) ([]models.UserTransaction, error)

	// Получение временных данных о транзакциях
	GetTemporalTransactionsByEventID(ctx context.Context, eventID int64) ([]models.EventTransactionTemporalResponse, error)
}

// TransactionTypeRepository представляет интерфейс для работы с типами транзакций
type TransactionTypeRepository interface {
	GetTransactionTypes(ctx context.Context) ([]models.TransactionType, error)
	GetTransactionTypeByID(ctx context.Context, id int) (models.TransactionType, error)
	CreateTransactionType(ctx context.Context, transactionType models.TransactionType) (models.TransactionType, error)
	UpdateTransactionType(ctx context.Context, transactionType models.TransactionType) error
	DeleteTransactionType(ctx context.Context, id int) error
}

// IconRepository представляет интерфейс для работы с иконками
type IconRepository interface {
	GetIcons(ctx context.Context) ([]models.Icon, error)
	GetIconByID(ctx context.Context, id string) (models.Icon, error)
	CreateIcon(ctx context.Context, icon models.Icon) (models.Icon, error)
	UpdateIcon(ctx context.Context, icon models.Icon) error
	DeleteIcon(ctx context.Context, id string) error
}
