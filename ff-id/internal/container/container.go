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
	"github.com/ivasnev/FinFlow/ff-id/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-id/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	avatarRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/avatar"
	friendRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/friend"
	userRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/user"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	friendService "github.com/ivasnev/FinFlow/ff-id/internal/service/friend"
	userService "github.com/ivasnev/FinFlow/ff-id/internal/service/user"
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmmiddleware "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB

	// Репозитории
	UserRepository   repository.User
	AvatarRepository repository.Avatar
	FriendRepository repository.Friend

	// Сервисы
	UserService   service.UserServiceInterface
	FriendService service.FriendServiceInterface

	// Обработчики
	ServerHandler *handler.ServerHandler

	// Клиенты
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

	// Инициализируем репозитории
	container.initRepositories()

	// Инициализируем сервисы
	container.initServices()

	// Инициализируем обработчики
	container.initHandlers()

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

	return container, nil
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

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() {
	c.UserRepository = userRepo.NewUserRepository(c.DB)
	c.AvatarRepository = avatarRepo.NewAvatarRepository(c.DB)
	c.FriendRepository = friendRepo.NewFriendRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.UserService = userService.NewUserService(c.UserRepository, c.AvatarRepository)
	c.FriendService = friendService.NewFriendService(c.FriendRepository, c.UserRepository)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.ServerHandler = handler.NewServerHandler(c.FriendService, c.UserService)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// API версии v1
	v1 := c.Router.Group("")

	// Middleware для авторизации
	authMiddleware := auth.AuthMiddleware(c.AuthClient)

	// Middleware для TVM
	tvmMiddleware := tvmmiddleware.NewTVMMiddleware(c.TVMClient)

	// Регистрируем маршруты с помощью сгенерированного сервера
	api.RegisterHandlersWithOptions(v1, c.ServerHandler, api.GinServerOptions{
		Middlewares: []api.MiddlewareFunc{
			func(c *gin.Context) {
				// Применяем middleware в зависимости от типа запроса
				if scopes, ok := c.Get(api.BearerAuthScopes); ok && scopes != nil {
					authMiddleware(c)
				} else if scopes, ok := c.Get(api.TVMAuthScopes); ok && scopes != nil {
					tvmMiddleware.ValidateTicket()(c)
				}
			},
		},
	})
}
