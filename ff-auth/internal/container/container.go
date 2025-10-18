package container

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/middleware"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	pg_repos "github.com/ivasnev/FinFlow/ff-auth/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	authService "github.com/ivasnev/FinFlow/ff-auth/internal/service/auth"
	deviceService "github.com/ivasnev/FinFlow/ff-auth/internal/service/device"
	loginHistoryService "github.com/ivasnev/FinFlow/ff-auth/internal/service/login_history"
	sessionService "github.com/ivasnev/FinFlow/ff-auth/internal/service/session"
	tokenService "github.com/ivasnev/FinFlow/ff-auth/internal/service/token"
	userService "github.com/ivasnev/FinFlow/ff-auth/internal/service/user"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	idclient "github.com/ivasnev/FinFlow/ff-id/pkg/client"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB

	// Репозитории
	UserRepository         pg_repos.UserRepositoryInterface
	RoleRepository         pg_repos.RoleRepositoryInterface
	SessionRepository      pg_repos.SessionRepositoryInterface
	LoginHistoryRepository pg_repos.LoginHistoryRepositoryInterface
	DeviceRepository       pg_repos.DeviceRepositoryInterface
	KeyPairRepository      pg_repos.KeyPairRepositoryInterface

	// Токен менеджер
	TokenManager service.TokenManager
	IDClient     *idclient.Client

	// Сервисы
	AuthService         service.Auth
	UserService         service.User
	SessionService      service.Session
	LoginHistoryService service.LoginHistory
	DeviceService       service.Device

	// Обработчики
	ServerHandler *handler.ServerHandler
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

	// Инициализируем TokenManager
	tokenManager, err := tokenService.NewED25519TokenManager(container.KeyPairRepository)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации менеджера токенов: %w", err)
	}
	container.TokenManager = tokenManager

	tvmClient := tvmclient.NewTVMClient(cfg.TVM.BaseURL, cfg.TVM.ServiceSecret)

	idClient := idclient.NewClient(cfg.IDClient.BaseURL, cfg.TVM.ServiceID, cfg.IDClient.TVMID, tvmClient)
	container.IDClient = idClient

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
	c.KeyPairRepository = pg_repos.NewKeyPairRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.DeviceService = deviceService.NewDeviceService(c.DeviceRepository)
	c.AuthService = authService.NewAuthService(
		c.Config,
		c.UserRepository,
		c.RoleRepository,
		c.SessionRepository,
		c.DeviceService,
		c.LoginHistoryRepository,
		c.TokenManager,
		c.IDClient,
	)
	c.UserService = userService.NewUserService(c.UserRepository)
	c.SessionService = sessionService.NewSessionService(c.SessionRepository)
	c.LoginHistoryService = loginHistoryService.NewLoginHistoryService(c.LoginHistoryRepository)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.ServerHandler = handler.NewServerHandler(
		c.AuthService,
		c.UserService,
		c.SessionService,
		c.LoginHistoryService,
		c.TokenManager,
	)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// Добавляем CORS middleware глобально для всех маршрутов
	c.Router.Use(middleware.CORSMiddleware())

	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Создаем адаптер для TokenManager
	tokenAdapter := &TokenManagerAdapter{tokenManager: c.TokenManager}

	// Middleware для авторизации
	authMiddleware := auth.AuthMiddleware(tokenAdapter)

	// Регистрируем маршруты с помощью сгенерированного сервера
	api.RegisterHandlersWithOptions(v1, c.ServerHandler, api.GinServerOptions{
		Middlewares: []api.MiddlewareFunc{func(c *gin.Context) { authMiddleware(c) }},
	})
}

// TokenManagerAdapter адаптирует service.TokenManager к auth.ValidateClient
type TokenManagerAdapter struct {
	tokenManager service.TokenManager
}

// ValidateToken реализует auth.ValidateClient
func (a *TokenManagerAdapter) ValidateToken(tokenStr string) (*auth.TokenPayload, error) {
	payload, err := a.tokenManager.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	// Конвертируем service.TokenPayload в auth.TokenPayload
	return &auth.TokenPayload{
		UserID: payload.UserID,
		Roles:  payload.Roles,
		Exp:    payload.Exp,
	}, nil
}
