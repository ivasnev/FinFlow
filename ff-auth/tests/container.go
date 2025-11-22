package tests

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/container"
	deviceRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/device"
	keyPairRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/key_pair"
	loginHistoryRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/login_history"
	roleRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/role"
	sessionRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/session"
	userRepository "github.com/ivasnev/FinFlow/ff-auth/internal/repository/user"
	authService "github.com/ivasnev/FinFlow/ff-auth/internal/service/auth"
	deviceService "github.com/ivasnev/FinFlow/ff-auth/internal/service/device"
	loginHistoryService "github.com/ivasnev/FinFlow/ff-auth/internal/service/login_history"
	sessionService "github.com/ivasnev/FinFlow/ff-auth/internal/service/session"
	tokenService "github.com/ivasnev/FinFlow/ff-auth/internal/service/token"
	userService "github.com/ivasnev/FinFlow/ff-auth/internal/service/user"
	"gorm.io/gorm"
)

// createTestContainer создает тестовый контейнер с роутером для HTTP сервера
func createTestContainer(t *testing.T, cfg *config.Config, router *gin.Engine, db *gorm.DB, idClient *ffid.Adapter) (*container.Container, error) {
	c := &container.Container{
		Config: cfg,
		Router: router,
		DB:     db,
	}

	// Инициализируем репозитории (копируем логику из container.initRepositories)
	c.UserRepository = userRepository.NewUserRepository(c.DB)
	c.RoleRepository = roleRepository.NewRoleRepository(c.DB)
	c.SessionRepository = sessionRepository.NewSessionRepository(c.DB)
	c.LoginHistoryRepository = loginHistoryRepository.NewLoginHistoryRepository(c.DB)
	c.DeviceRepository = deviceRepository.NewDeviceRepository(c.DB)
	c.KeyPairRepository = keyPairRepository.NewKeyPairRepository(c.DB)

	// Инициализируем TokenManager (копируем логику из container.NewContainer)
	tokenManager, err := tokenService.NewED25519TokenManager(c.KeyPairRepository)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации менеджера токенов: %w", err)
	}
	c.TokenManager = tokenManager

	// Устанавливаем переданный ID клиент (вместо создания с TVM транспортом)
	c.IDClient = idClient

	// Инициализируем сервисы (копируем логику из container.initServices)
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

	// Инициализируем обработчики (копируем логику из container.initHandlers)
	c.ServerHandler = handler.NewServerHandler(
		c.AuthService,
		c.UserService,
		c.SessionService,
		c.LoginHistoryService,
		c.TokenManager,
	)

	return c, nil
}
