package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Settlement хранит журнал возвратов и сериализует его изменения внутри мероприятия.
type Settlement interface {
	// InEvent выполняет callback под SQL-замком строки мероприятия.
	InEvent(context.Context, int64, func(context.Context) error) error
	// Members возвращает внутренние ID и внешние ID действующих участников.
	Members(context.Context, int64) (map[int64]*int64, error)
	// Ledger возвращает расходные и обратные возвратные рёбра в одном снимке.
	Ledger(context.Context, int64) ([]models.Debt, error)
	// ByRequest ищет также отменённую запись для идемпотентного повтора.
	ByRequest(context.Context, int64, int64, string) (*models.Settlement, error)
	// Get ищет запись только в указанном мероприятии.
	Get(context.Context, int64, int64) (*models.Settlement, error)
	// List читает до 50 записей до указанного ID.
	List(context.Context, int64, int64) ([]models.Settlement, error)
	// Create сохраняет возврат и инвалидирует версию оптимизации в общей транзакции.
	Create(context.Context, *models.Settlement) error
	// Cancel помечает возврат отменённым и инвалидирует версию оптимизации.
	Cancel(context.Context, int64, int64) error
}
