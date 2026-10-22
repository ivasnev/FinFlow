package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Receipt владеет атомарными переходами и короткой арендой фоновых задач.
type Receipt interface {
	// InEvent сериализует изменение чека с изменениями события.
	InEvent(context.Context, int64, func(context.Context) error) error
	// IsMember проверяет участие по внешнему ID пользователя.
	IsMember(context.Context, int64, int64) (bool, error)
	// Create резервирует загрузку после дедупликации и проверки квот.
	Create(context.Context, *models.Receipt) (*models.Receipt, error)
	// Get читает чек своего события с опциональной блокировкой.
	Get(context.Context, int64, string, bool) (*models.Receipt, error)
	// List возвращает ограниченную страницу чеков.
	List(context.Context, int64, int64) (*models.ReceiptList, error)
	// Save сохраняет переход с проверкой ревизии.
	Save(context.Context, *models.Receipt) error
	// Retry сохраняет ручной повтор с отдельным учётом квот.
	Retry(context.Context, *models.Receipt, int64) error
	// Remove скрывает неподтверждённый чек и сохраняет outbox удаления.
	Remove(context.Context, *models.Receipt) error
	// Claim выдаёт задачу одному обработчику на ограниченное время.
	Claim(context.Context) (*models.Receipt, error)
	// Complete принимает результат только от владельца действующей аренды.
	Complete(context.Context, *models.Receipt, *models.ReceiptResult, error) error
	// Garbage возвращает приватные объекты для отложенного удаления.
	Garbage(context.Context) ([]string, error)
	// DeleteGarbage подтверждает удаление объекта из хранилища.
	DeleteGarbage(context.Context, string) error
	// FinishUpload завершает резервирование после записи фотографии.
	FinishUpload(context.Context, *models.Receipt, bool) error
}
