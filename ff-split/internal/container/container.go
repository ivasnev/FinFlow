package container

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-id/pkg/client"
	handler "github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/middleware"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	pg_repos "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres"
	category_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/category"
	service "github.com/ivasnev/FinFlow/ff-split/internal/service"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"

	//tvmmiddleware "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Импорт сгенерированной Swagger-документации
	_ "github.com/ivasnev/FinFlow/ff-split/docs"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB

	// Репозитории
	CategoryRepository    *category_repository.Repository
	EventRepository       *pg_repos.EventRepository
	ActivityRepository    *pg_repos.ActivityRepository
	UserRepository        *pg_repos.UserRepository
	IconRepository        *pg_repos.IconRepository
	TaskRepository        *pg_repos.TaskRepository
	TransactionRepository *pg_repos.TransactionRepository

	// Сервисы
	CategoryService    service.CategoryServiceInterface
	EventService       service.EventServiceInterface
	ActivityService    service.ActivityServiceInterface
	UserService        service.UserServiceInterface
	IconService        service.IconServiceInterface
	TaskService        service.TaskServiceInterface
	TransactionService service.TransactionServiceInterface

	// Обработчики маршрутов
	CategoryHandler    handler.CategoryHandlerInterface
	EventHandler       handler.EventHandlerInterface
	ActivityHandler    handler.ActivityHandlerInterface
	IconHandler        handler.IconHandlerInterface
	TaskHandler        handler.TaskHandlerInterface
	TransactionHandler handler.TransactionHandlerInterface
	UserHandler        handler.UserHandlerInterface

	// Клиенты внешних сервисов
	AuthClient *auth.Client
	TVMClient  *tvmclient.TVMClient
	IDClient   *client.Client
}

// NewContainer - конструктор контейнера зависимостей
func NewContainer(cfg *config.Config, router *gin.Engine) (*Container, error) {
	container := &Container{
		Config: cfg,
		Router: router,
	}

	// Инициализируем подключение к базе данных
	if err := container.initDB(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации базы данных: %w", err)
	}

	// Инициализируем клиенты
	container.AuthClient = auth.NewClient(
		cfg.AuthClient.Host+":"+strconv.Itoa(cfg.AuthClient.Port),
		time.Second*time.Duration(cfg.AuthClient.UpdateInterval),
	)

	// Инициализируем TVM клиент
	container.TVMClient = tvmclient.NewTVMClient(
		cfg.TVM.BaseURL,
		cfg.TVM.ServiceSecret,
	)

	// Инициализируем ID клиент
	container.IDClient = client.NewClient(
		cfg.IDService.BaseURL,
		cfg.TVM.ServiceID,
		cfg.IDService.ServiceID,
		container.TVMClient,
	)

	container.initRepositories()
	container.initServices()
	container.initHandlers()

	return container, nil
}

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() {
	c.CategoryRepository = category_repository.NewRepository(c.DB)
	c.EventRepository = pg_repos.NewEventRepository(c.DB)
	c.ActivityRepository = pg_repos.NewActivityRepository(c.DB)
	c.UserRepository = pg_repos.NewUserRepository(c.DB)
	c.IconRepository = pg_repos.NewIconRepository(c.DB)
	c.TaskRepository = pg_repos.NewTaskRepository(c.DB)
	c.TransactionRepository = pg_repos.NewTransactionRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.UserService = service.NewUserService(c.UserRepository, c.IDClient)
	c.CategoryService = service.NewCategoryService(c.CategoryRepository)
	c.EventService = service.NewEventService(c.EventRepository, c.DB, c.UserService)
	c.ActivityService = service.NewActivityService(c.ActivityRepository)
	c.IconService = service.NewIconService(c.IconRepository)
	c.TaskService = service.NewTaskService(c.TaskRepository, c.UserService)
	c.TransactionService = service.NewTransactionService(c.DB, c.TransactionRepository, c.UserService, c.EventService)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.CategoryHandler = handler.NewCategoryHandler(c.CategoryService)
	c.EventHandler = handler.NewEventHandler(c.EventService, c.UserService)
	c.ActivityHandler = handler.NewActivityHandler(c.ActivityService)
	c.IconHandler = handler.NewIconHandler(c.IconService)
	c.TaskHandler = handler.NewTaskHandler(c.TaskService)
	c.TransactionHandler = handler.NewTransactionHandler(c.TransactionService)
	c.UserHandler = handler.NewUserHandler(c.UserService)
}

// initDB инициализирует подключение к базе данных
func (c *Container) initDB() error {
	// Формируем строку подключения к PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Config.Postgres.Host,
		c.Config.Postgres.Port,
		c.Config.Postgres.User,
		c.Config.Postgres.Password,
		c.Config.Postgres.DBName,
	)

	// Определяем уровень логирования на основе конфигурации
	var logLevel logger.LogLevel
	switch c.Config.Logger.Level {
	case config.LogLevelSilent:
		logLevel = logger.Silent
	case config.LogLevelError:
		logLevel = logger.Error
	case config.LogLevelWarn:
		logLevel = logger.Warn
	case config.LogLevelInfo:
		logLevel = logger.Info
	default:
		logLevel = logger.Info // По умолчанию используем Info
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logLevel,    // Log level from config
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// Добавляем CORS middleware глобально для всех маршрутов
	c.Router.Use(middleware.CORSMiddleware())

	// Swagger
	c.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Middleware для авторизации
	authMiddleware := auth.AuthMiddleware(c.AuthClient)

	// Middleware для TVM
	//tvmMiddleware := tvmmiddleware.NewTVMMiddleware(c.TVMClient)

	// Категории
	categoryRoutes := v1.Group("/category")
	{
		categoryRoutes.OPTIONS("", c.CategoryHandler.Options)
		categoryRoutes.GET("", c.CategoryHandler.GetCategories)
		categoryRoutes.GET("/:id", c.CategoryHandler.GetCategoryByID)
	}

	// Мероприятия
	eventRoutes := v1.Group("/event", authMiddleware)
	{
		// Список мероприятий
		eventRoutes.GET("/", c.EventHandler.GetEvents)

		// Маршруты для отдельного мероприятия
		eventRoutes.GET("/:id_event", c.EventHandler.GetEventByID)
		eventRoutes.POST("", c.EventHandler.CreateEvent)
		eventRoutes.PUT("/:id_event", c.EventHandler.UpdateEvent)
		eventRoutes.DELETE("/:id_event", c.EventHandler.DeleteEvent)

		activityRoutes := eventRoutes.Group("/:id_event/activity")
		{
			// Активности мероприятия
			activityRoutes.GET("", c.ActivityHandler.GetActivitiesByEventID)
			activityRoutes.GET("/:id_activity", c.ActivityHandler.GetActivityByID)
			activityRoutes.POST("", c.ActivityHandler.CreateActivity)
			activityRoutes.PUT("/:id_activity", c.ActivityHandler.UpdateActivity)
			activityRoutes.DELETE("/:id_activity", c.ActivityHandler.DeleteActivity)
		}

		// Задачи мероприятия
		taskRoutes := eventRoutes.Group("/:id_event/task")
		{
			taskRoutes.GET("", c.TaskHandler.GetTasksByEventID)
			taskRoutes.GET("/:id_task", c.TaskHandler.GetTaskByID)
			taskRoutes.POST("", c.TaskHandler.CreateTask)
			taskRoutes.PUT("/:id_task", c.TaskHandler.UpdateTask)
			taskRoutes.DELETE("/:id_task", c.TaskHandler.DeleteTask)
		}

		// Транзакции мероприятия
		transactionRoutes := eventRoutes.Group("/:id_event/transaction")
		{
			transactionRoutes.GET("", c.TransactionHandler.GetTransactionsByEventID)
			transactionRoutes.GET("/:id_transaction", c.TransactionHandler.GetTransactionByID)
			transactionRoutes.POST("", c.TransactionHandler.CreateTransaction)
			transactionRoutes.PUT("/:id_transaction", c.TransactionHandler.UpdateTransaction)
			transactionRoutes.DELETE("/:id_transaction", c.TransactionHandler.DeleteTransaction)
		}

		// Долги мероприятия
		eventRoutes.GET("/:id_event/debts", c.TransactionHandler.GetDebtsByEventID)

		// Оптимизированные долги мероприятия
		eventRoutes.GET("/:id_event/optimized-debts", c.TransactionHandler.GetOptimizedDebtsByEventID)
		eventRoutes.POST("/:id_event/optimized-debts", c.TransactionHandler.OptimizeDebts)
		eventRoutes.GET("/:id_event/user/:id_user/optimized-debts", c.TransactionHandler.GetOptimizedDebtsByUserID)

		// Пользователи мероприятия
		users := eventRoutes.Group("/:id_event/user")
		{
			users.GET("", c.UserHandler.GetUsersByEventID)
			users.POST("", c.UserHandler.AddUsersToEvent)
			users.DELETE("/:id_user", c.UserHandler.RemoveUserFromEvent)

			// Dummy-пользователи
			users.GET("/dummies", c.UserHandler.GetDummiesByEventID)
			users.POST("/dummy", c.UserHandler.CreateDummyUser)
			users.POST("/dummies", c.UserHandler.BatchCreateDummyUsers)
		}
	}

	// Управление (требуется роль service_admin)
	manageRoutes := v1.Group("/manage")
	{
		categoryManageRoutes := manageRoutes.Group("/category")
		{
			categoryManageRoutes.OPTIONS("", c.CategoryHandler.Options)
			categoryManageRoutes.POST("", c.CategoryHandler.CreateCategory)
			categoryManageRoutes.PUT("/:id", c.CategoryHandler.UpdateCategory)
			categoryManageRoutes.DELETE("/:id", c.CategoryHandler.DeleteCategory)
		}
		// Типы транзакций
		manageRoutes.Group("/transaction_type")
		{
			// Здесь будут добавлены маршруты для типов транзакций
		}

		// Иконки
		iconsRoutes := manageRoutes.Group("/icons")
		{
			iconsRoutes.GET("", c.IconHandler.GetIcons)
			iconsRoutes.GET("/:id", c.IconHandler.GetIconByID)
			iconsRoutes.POST("", c.IconHandler.CreateIcon)
			iconsRoutes.PUT("/:id", c.IconHandler.UpdateIcon)
			iconsRoutes.DELETE("/:id", c.IconHandler.DeleteIcon)
		}
	}

	// Пользователи
	users := v1.Group("/user")
	{
		users.GET("/:id_user", c.UserHandler.GetUserByID)
		users.PUT("/:id_user", c.UserHandler.UpdateUser)
		users.DELETE("/:id_user", c.UserHandler.DeleteUser)
		users.POST("/sync", c.UserHandler.SyncUsers)
		users.POST("/list", c.UserHandler.GetUsersByIDs)
	}

	// Базовый маршрут для проверки работоспособности сервиса
	v1.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "ok",
			"name":   "FinFlow Split Service",
		})
	})
}
