package container

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/middleware"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	deviceRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/device"
	keyPairRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/key_pair"
	loginHistoryRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/login_history"
	roleRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/role"
	sessionRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/session"
	userRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/user"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	authService "github.com/ivasnev/FinFlow/ff-auth/internal/service/auth"
	deviceService "github.com/ivasnev/FinFlow/ff-auth/internal/service/device"
	loginHistoryService "github.com/ivasnev/FinFlow/ff-auth/internal/service/login_history"
	sessionService "github.com/ivasnev/FinFlow/ff-auth/internal/service/session"
	tokenService "github.com/ivasnev/FinFlow/ff-auth/internal/service/token"
	userService "github.com/ivasnev/FinFlow/ff-auth/internal/service/user"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmtransport "github.com/ivasnev/FinFlow/ff-tvm/pkg/transport"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config *config.Config
	Router *gin.Engine
	DB     *gorm.DB

	// Репозитории
	UserRepository         repository.User
	RoleRepository         repository.Role
	SessionRepository      repository.Session
	LoginHistoryRepository repository.LoginHistory
	DeviceRepository       repository.Device
	KeyPairRepository      repository.KeyPair

	// Токен менеджер
	TokenManager service.TokenManager
	IDClient     *ffid.Adapter

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

	// Создаем TVM транспорт для ff-id клиента
	tvmTransport := tvmtransport.NewTVMTransport(tvmClient, http.DefaultTransport, cfg.TVM.ServiceID, cfg.IDClient.TVMID)
	httpClient := &http.Client{
		Transport: tvmTransport,
		Timeout:   10 * time.Second,
	}

	idClient, err := ffid.NewAdapter(cfg.IDClient.BaseURL, httpClient)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации ID клиента: %w", err)
	}
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
	c.UserRepository = userRepository.NewUserRepository(c.DB)
	c.RoleRepository = roleRepository.NewRoleRepository(c.DB)
	c.SessionRepository = sessionRepository.NewSessionRepository(c.DB)
	c.LoginHistoryRepository = loginHistoryRepository.NewLoginHistoryRepository(c.DB)
	c.DeviceRepository = deviceRepository.NewDeviceRepository(c.DB)
	c.KeyPairRepository = keyPairRepository.NewKeyPairRepository(c.DB)
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
		Middlewares: []api.MiddlewareFunc{
			func(c *gin.Context) {
				// Применяем middleware в зависимости от типа запроса
				if scopes, ok := c.Get(api.BearerAuthScopes); ok && scopes != nil {
					authMiddleware(c)
				}
			},
		},
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
