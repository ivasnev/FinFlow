package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Service представляет основной интерфейс сервиса
type Service interface {
	CategoryService() CategoryService
	EventService() EventService
	ActivityService() ActivityService
	TransactionService() TransactionService
	TransactionTypeService() TransactionTypeService
	IconService() IconService
	FileService() FileService
}

// CategoryService представляет интерфейс сервиса для работы с категориями
type CategoryService interface {
	GetCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	GetCategoryByID(ctx context.Context, id int) (dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, category models.Category) (dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, category models.Category) error
	DeleteCategory(ctx context.Context, id int) error
}

// EventService представляет интерфейс сервиса для работы с мероприятиями
type EventService interface {
	GetEvents(ctx context.Context) ([]dto.EventResponse, error)
	GetEventByID(ctx context.Context, id int64) (dto.EventResponse, error)
	CreateEvent(ctx context.Context, eventRequest dto.EventRequest) (dto.EventResponse, error)
	UpdateEvent(ctx context.Context, id int64, eventRequest dto.EventRequest) error
	DeleteEvent(ctx context.Context, id int64) error
}

// ActivityService представляет интерфейс сервиса для работы с активностями
type ActivityService interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]dto.ActivityResponse, error)
	GetActivityByID(ctx context.Context, id int) (dto.ActivityResponse, error)
	CreateActivity(ctx context.Context, activity models.Activity) (dto.ActivityResponse, error)
	UpdateActivity(ctx context.Context, activity models.Activity) error
	DeleteActivity(ctx context.Context, id int) error
}

// TransactionService представляет интерфейс сервиса для работы с транзакциями
type TransactionService interface {
	GetTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.TransactionResponse, error)
	GetTransactionByID(ctx context.Context, id int) (dto.TransactionResponse, error)
	CreateTransaction(ctx context.Context, transaction models.Transaction, userIDs []int64) (dto.TransactionResponse, error)
	UpdateTransaction(ctx context.Context, transaction models.Transaction, userIDs []int64) error
	DeleteTransaction(ctx context.Context, id int) error
	GetTemporalTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.EventTransactionTemporalResponse, error)
}

// TransactionTypeService представляет интерфейс сервиса для работы с типами транзакций
type TransactionTypeService interface {
	GetTransactionTypes(ctx context.Context) ([]models.TransactionType, error)
	GetTransactionTypeByID(ctx context.Context, id int) (models.TransactionType, error)
	CreateTransactionType(ctx context.Context, transactionType models.TransactionType) (models.TransactionType, error)
	UpdateTransactionType(ctx context.Context, transactionType models.TransactionType) error
	DeleteTransactionType(ctx context.Context, id int) error
}

// IconService представляет интерфейс сервиса для работы с иконками
type IconService interface {
	GetIcons(ctx context.Context) ([]models.Icon, error)
	GetIconByID(ctx context.Context, id string) (models.Icon, error)
	CreateIcon(ctx context.Context, icon models.Icon) (models.Icon, error)
	UpdateIcon(ctx context.Context, icon models.Icon) error
	DeleteIcon(ctx context.Context, id string) error
}

// FileService представляет интерфейс сервиса для работы с файлами
type FileService interface {
	UploadFile(ctx context.Context, fileName string, fileData []byte, contentType string) (string, error)
	GetFile(ctx context.Context, fileUUID string) ([]byte, string, error)
	DeleteFile(ctx context.Context, fileUUID string) error
}
