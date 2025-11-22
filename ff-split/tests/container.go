package tests

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/container"
	activity_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/activity"
	category_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/category"
	event_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/event"
	icon_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/icon"
	task_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/task"
	transaction_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/transaction"
	user_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/user"
	activity_service "github.com/ivasnev/FinFlow/ff-split/internal/service/activity"
	category_service "github.com/ivasnev/FinFlow/ff-split/internal/service/category"
	event_service "github.com/ivasnev/FinFlow/ff-split/internal/service/event"
	icon_service "github.com/ivasnev/FinFlow/ff-split/internal/service/icon"
	task_service "github.com/ivasnev/FinFlow/ff-split/internal/service/task"
	transaction_service "github.com/ivasnev/FinFlow/ff-split/internal/service/transaction"
	user_service "github.com/ivasnev/FinFlow/ff-split/internal/service/user"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createTestContainer создает тестовый контейнер с роутером для HTTP сервера
func createTestContainer(t *testing.T, cfg *config.Config, router *gin.Engine, db *gorm.DB, httpClient *http.Client) (*container.Container, error) {
	c := &container.Container{
		Config: cfg,
		Router: router,
		DB:     db,
	}

	// Инициализируем репозитории
	c.CategoryRepository = category_repository.NewRepository(c.DB)
	c.EventRepository = event_repository.NewEventRepository(c.DB)
	c.ActivityRepository = activity_repository.NewActivityRepository(c.DB)
	c.UserRepository = user_repository.NewUserRepository(c.DB)
	c.IconRepository = icon_repository.NewIconRepository(c.DB)
	c.TaskRepository = task_repository.NewTaskRepository(c.DB)
	c.TransactionRepository = transaction_repository.NewTransactionRepository(c.DB)

	// Создаем реальный HTTP адаптер для ff-id (будет использовать MockServer)
	idAdapter, err := ffid.NewAdapter(cfg.IDService.BaseURL, httpClient)
	require.NoError(t, err, "не удалось создать ff-id адаптер")

	// Инициализируем сервисы с реальным HTTP адаптером
	c.UserService = user_service.NewUserService(c.UserRepository, idAdapter)
	c.CategoryService = category_service.NewCategoryService(c.CategoryRepository)
	c.EventService = event_service.NewEventService(c.EventRepository, c.DB, c.UserService, c.CategoryService)
	c.ActivityService = activity_service.NewActivityService(c.ActivityRepository)
	c.IconService = icon_service.NewIconService(c.IconRepository)
	c.TaskService = task_service.NewTaskService(c.TaskRepository, c.UserService)
	c.TransactionService = transaction_service.NewTransactionService(c.DB, c.TransactionRepository, c.UserService, c.EventService)

	return c, nil
}
