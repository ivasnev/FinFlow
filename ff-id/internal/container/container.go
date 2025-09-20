package container

import (
	"fmt"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-id/internal/common/config"
	pg_repos "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB

	// Репозитории
	UserRepository         interfaces.UserRepositoryInterface
	RoleRepository         interfaces.RoleRepositoryInterface
	SessionRepository      interfaces.SessionRepositoryInterface
	LoginHistoryRepository interfaces.LoginHistoryRepositoryInterface
	DeviceRepository       interfaces.DeviceRepositoryInterface
	AvatarRepository       interfaces.AvatarRepositoryInterface

	// Сервисы
	AuthService         service.AuthServiceInterface
	UserService         service.UserServiceInterface
	SessionService      service.SessionServiceInterface
	LoginHistoryService service.LoginHistoryServiceInterface
	DeviceService       service.DeviceServiceInterface

	// Обработчики
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	SessionHandler *handler.SessionHandler
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

	// Подключаемся к базе данных
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	c.DB = db
	return nil
}

// initRepositories инициализирует репозитории
func (c *Container) initRepositories() {
	c.UserRepository = pg_repos.NewUserRepository(c.DB)
	c.RoleRepository = pg_repos.NewRoleRepository(c.DB)
	c.SessionRepository = pg_repos.NewSessionRepository(c.DB)
	c.LoginHistoryRepository = pg_repos.NewLoginHistoryRepository(c.DB)
	c.DeviceRepository = pg_repos.NewDeviceRepository(c.DB)
	c.AvatarRepository = pg_repos.NewAvatarRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.DeviceService = service.NewDeviceService(c.DeviceRepository)
	c.AuthService = service.NewAuthService(
		c.Config,
		c.UserRepository,
		c.RoleRepository,
		c.SessionRepository,
		c.DeviceService,
		c.LoginHistoryRepository,
	)
	c.UserService = service.NewUserService(c.UserRepository, c.AvatarRepository)
	c.SessionService = service.NewSessionService(c.SessionRepository)
	c.LoginHistoryService = service.NewLoginHistoryService(c.LoginHistoryRepository)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.AuthHandler = handler.NewAuthHandler(c.AuthService)
	c.UserHandler = handler.NewUserHandler(c.UserService)
	c.SessionHandler = handler.NewSessionHandler(c.SessionService, c.LoginHistoryService)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Middleware для авторизации
	authMiddleware := middleware.AuthMiddleware(c.AuthService)

	// Группа маршрутов для аутентификации
	auth := v1.Group("/auth")
	{
		// Публичные маршруты
		auth.POST("/register", c.AuthHandler.Register)
		auth.POST("/login", c.AuthHandler.Login)
		auth.POST("/refresh", c.AuthHandler.RefreshToken)

		// Защищенные маршруты
		auth.Use(authMiddleware)
		auth.POST("/logout", c.AuthHandler.Logout)
	}

	// Группа маршрутов для пользователей
	users := v1.Group("/users")
	{
		// Публичные маршруты
		users.GET("/:nickname", c.UserHandler.GetUserByNickname)

		// Защищенные маршруты
		users.Use(authMiddleware)
		users.PATCH("/me", c.UserHandler.UpdateUser)
	}

	// Группа маршрутов для сессий (все требуют аутентификации)
	sessions := v1.Group("/sessions", authMiddleware)
	{
		sessions.GET("", c.SessionHandler.GetUserSessions)
		sessions.DELETE("/:id", c.SessionHandler.TerminateSession)
	}

	// Группа маршрутов для истории входов (все требуют аутентификации)
	loginHistory := v1.Group("/login-history", authMiddleware)
	{
		loginHistory.GET("", c.SessionHandler.GetLoginHistory)
	}
}
