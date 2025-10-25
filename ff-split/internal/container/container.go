package container

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	ffidadapter "github.com/ivasnev/FinFlow/ff-split/internal/adapters/ffid"
	handler "github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/middleware"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	activity_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/activity"
	category_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/category"
	event_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/event"
	icon_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/icon"
	task_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/task"
	transaction_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/transaction"
	user_repository "github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/user"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	activity_service "github.com/ivasnev/FinFlow/ff-split/internal/service/activity"
	category_service "github.com/ivasnev/FinFlow/ff-split/internal/service/category"
	event_service "github.com/ivasnev/FinFlow/ff-split/internal/service/event"
	icon_service "github.com/ivasnev/FinFlow/ff-split/internal/service/icon"
	task_service "github.com/ivasnev/FinFlow/ff-split/internal/service/task"
	transaction_service "github.com/ivasnev/FinFlow/ff-split/internal/service/transaction"
	user_service "github.com/ivasnev/FinFlow/ff-split/internal/service/user"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmtransport "github.com/ivasnev/FinFlow/ff-tvm/pkg/transport"

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
	CategoryRepository    repository.Category
	EventRepository       repository.Event
	ActivityRepository    repository.Activity
	UserRepository        repository.User
	IconRepository        repository.Icon
	TaskRepository        repository.Task
	TransactionRepository repository.Transaction

	// Сервисы
	CategoryService    service.Category
	EventService       service.Event
	ActivityService    service.Activity
	UserService        service.User
	IconService        service.Icon
	TaskService        service.Task
	TransactionService service.Transaction

	// Адаптеры
	IDAdapter *ffidadapter.Adapter

	// Обработчик
	ServerHandler *handler.ServerHandler

	// Клиенты внешних сервисов
	AuthClient *auth.Client
	TVMClient  *tvmclient.TVMClient
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

	// Инициализируем TVM транспорт для ff-id
	tvmTransport := tvmtransport.NewTVMTransport(
		container.TVMClient,
		http.DefaultTransport,
		cfg.TVM.ServiceID,
		cfg.IDService.ServiceID,
	)
	httpClient := &http.Client{
		Transport: tvmTransport,
		Timeout:   10 * time.Second,
	}

	// Инициализируем адаптер ff-id
	var err error
	container.IDAdapter, err = ffidadapter.NewAdapter(cfg.IDService.BaseURL, httpClient)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации ff-id адаптера: %w", err)
	}

	container.initRepositories()
	container.initServices()
	container.initHandler()

	return container, nil
}

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() {
	c.CategoryRepository = category_repository.NewRepository(c.DB)
	c.EventRepository = event_repository.NewEventRepository(c.DB)
	c.ActivityRepository = activity_repository.NewActivityRepository(c.DB)
	c.UserRepository = user_repository.NewUserRepository(c.DB)
	c.IconRepository = icon_repository.NewIconRepository(c.DB)
	c.TaskRepository = task_repository.NewTaskRepository(c.DB)
	c.TransactionRepository = transaction_repository.NewTransactionRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.UserService = user_service.NewUserService(c.UserRepository, c.IDAdapter)
	c.CategoryService = category_service.NewCategoryService(c.CategoryRepository)
	c.EventService = event_service.NewEventService(c.EventRepository, c.DB, c.UserService, c.CategoryService)
	c.ActivityService = activity_service.NewActivityService(c.ActivityRepository)
	c.IconService = icon_service.NewIconService(c.IconRepository)
	c.TaskService = task_service.NewTaskService(c.TaskRepository, c.UserService)
	c.TransactionService = transaction_service.NewTransactionService(c.DB, c.TransactionRepository, c.UserService, c.EventService)
}

// initHandler инициализирует ServerHandler
func (c *Container) initHandler() {
	c.ServerHandler = handler.NewServerHandler(
		c.EventService,
		c.UserService,
		c.TransactionService,
		c.ActivityService,
		c.TaskService,
		c.CategoryService,
		c.IconService,
	)
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

	// Middleware для авторизации
	authMiddleware := auth.AuthMiddleware(c.AuthClient)

	// Регистрируем маршруты с помощью сгенерированного кода
	api.RegisterHandlersWithOptions(c.Router, c.ServerHandler, api.GinServerOptions{
		BaseURL: "",
		Middlewares: []api.MiddlewareFunc{
			func(c *gin.Context) {
				authMiddleware(c)
			},
		},
	})

	// Базовый маршрут для проверки работоспособности сервиса
	c.Router.GET("/api/v1/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "ok",
			"name":   "FinFlow Split Service",
		})
	})
}
