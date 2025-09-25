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
	pg_repos "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
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
	UserRepository   pg_repos.UserRepositoryInterface
	AvatarRepository pg_repos.AvatarRepositoryInterface
	FriendRepository pg_repos.FriendRepositoryInterface

	// Сервисы
	UserService   service.UserServiceInterface
	FriendService service.FriendServiceInterface

	// Обработчики
	UserHandler   *handler.UserHandler
	FriendHandler *handler.FriendHandler

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
	c.UserRepository = pg_repos.NewUserRepository(c.DB)
	c.AvatarRepository = pg_repos.NewAvatarRepository(c.DB)
	c.FriendRepository = pg_repos.NewFriendRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.UserService = service.NewUserService(c.UserRepository, c.AvatarRepository)
	c.FriendService = service.NewFriendService(c.FriendRepository, c.UserRepository)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.FriendHandler = handler.NewFriendHandler(c.FriendService)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Middleware для авторизации
	authMiddleware := auth.AuthMiddleware(c.AuthClient)

	// Middleware для TVM
	tvmMiddleware := tvmmiddleware.NewTVMMiddleware(c.TVMClient)

	// Группа маршрутов для пользователей
	users := v1.Group("/users")
	{
		// Публичные маршруты
		users.GET("/:nickname", c.UserHandler.GetUserByNickname)

		// Получение списка друзей
		users.GET("/:nickname/friends", c.FriendHandler.GetFriends)

		// Регистрация через внешний сервис
		users.POST("/register", authMiddleware, c.UserHandler.RegisterUser)

		// Защищенные маршруты
		users.Use(authMiddleware)
		users.PATCH("/me", c.UserHandler.UpdateUser)

		// Маршруты для управления друзьями
		users.POST("/me/friends", c.FriendHandler.AddFriend)
		users.DELETE("/me/friends/:friend_id", c.FriendHandler.RemoveFriend)

		// Маршруты для действий с заявками в друзья
		users.POST("/me/friends/action", c.FriendHandler.FriendAction)
		users.GET("/me/friend-requests", c.FriendHandler.GetFriendRequests)
	}

	// Внутренние маршруты для межсервисного взаимодействия
	internal := c.Router.Group("/internal")
	{
		// Защищенные TVM маршруты
		internalUsers := internal.Group("/users", tvmMiddleware.ValidateTicket())
		{
			// Регистрация через другой сервис (backend-to-backend)
			internalUsers.POST("/register", c.UserHandler.RegisterUserFromService)
		}
	}
}
