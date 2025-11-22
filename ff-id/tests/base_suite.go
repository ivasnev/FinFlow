package tests

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-id/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/container"
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	testTime "github.com/ivasnev/FinFlow/ff-id/tests/utils/time"
	testUUID "github.com/ivasnev/FinFlow/ff-id/tests/utils/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// BaseSuite представляет базовый набор для интеграционных тестов
type BaseSuite struct {
	suite.Suite
	Ctx         context.Context
	cancel      func()
	DBContainer *TestDBContainer
	HTTPClient  *http.Client
	Server      *httptest.Server
	Container   *container.Container
	Config      *config.Config
	APIClient   api.ClientWithResponsesInterface

	// Провайдеры для детерминированности
	TimeProvider *testTime.ConstantProvider
	UUIDProvider *testUUID.ConstantProvider
}

// SetupSuite выполняется один раз перед всеми тестами в suite
func (s *BaseSuite) SetupSuite() {
	// 1. Создание контекста
	s.Ctx, s.cancel = context.WithCancel(context.Background())

	// 2. Инициализация провайдеров для детерминированности
	s.TimeProvider = testTime.NewConstantProvider()
	s.UUIDProvider = testUUID.NewConstantProvider()

	// 3. Настройка тестовой БД с использованием testcontainers
	s.DBContainer = setupTestDB(s.T())

	// 4. Создаем конфигурацию для тестов
	s.Config = &config.Config{}
	s.Config.Postgres.Host = "localhost"
	s.Config.Postgres.Port = 5432
	s.Config.Postgres.User = "postgres"
	s.Config.Postgres.Password = "postgres"
	s.Config.Postgres.DBName = "ff_id_test"

	// 5. Создаем HTTP клиент для тестов
	s.HTTPClient = &http.Client{}

	// 6. Создаем Gin router в тестовом режиме
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 7. Создаем контейнер с тестовой БД
	var err error
	s.Container, err = createTestContainer(s.T(), s.Config, router, s.DBContainer.DB)
	require.NoError(s.T(), err, "не удалось создать тестовый контейнер")

	// 8. Создаем обработчики
	s.Container.ServerHandler = handler.NewServerHandler(s.Container.FriendService, s.Container.UserService)

	// 9. Регистрируем роуты (без middleware для упрощения тестирования)
	v1 := router.Group("/api/v1")
	api.RegisterHandlers(v1, s.Container.ServerHandler)

	// 10. Создаем тестовый HTTP сервер
	s.Server = httptest.NewServer(router)

	// 11. Создаем API клиент для удобных запросов
	apiClient, err := api.NewClientWithResponses(s.Server.URL+"/api/v1", api.WithHTTPClient(s.HTTPClient))
	require.NoError(s.T(), err, "не удалось создать API клиент")
	s.APIClient = apiClient
}

// TearDownSuite выполняется один раз после всех тестов в suite
func (s *BaseSuite) TearDownSuite() {
	// Закрытие HTTP сервера
	if s.Server != nil {
		s.Server.Close()
	}

	// Закрытие тестовой БД
	if s.DBContainer != nil {
		teardownTestDB(s.T(), s.DBContainer)
	}

	// Отмена контекста
	if s.cancel != nil {
		s.cancel()
	}
}

// SetupTest выполняется перед каждым тестом
func (s *BaseSuite) SetupTest() {
	// Очищаем данные перед каждым тестом для гарантированной чистоты
	s.cleanupDatabase()
}

// TearDownTest выполняется после каждого теста
func (s *BaseSuite) TearDownTest() {
	// Очищаем данные после каждого теста
	s.cleanupDatabase()
}

// cleanupDatabase очищает тестовую базу данных
func (s *BaseSuite) cleanupDatabase() {
	if s.DBContainer != nil && s.DBContainer.DB != nil {
		// Выполняем очистку в правильном порядке из-за внешних ключей
		s.DBContainer.DB.Exec("TRUNCATE TABLE user_friends CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE user_avatars CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE users CASCADE")
		// Сбрасываем последовательности
		s.DBContainer.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE user_friends_id_seq RESTART WITH 1")
	}
}

// GetDB возвращает подключение к тестовой БД
func (s *BaseSuite) GetDB() *gorm.DB {
	return s.DBContainer.DB
}

// GetServerURL возвращает URL тестового сервера
func (s *BaseSuite) GetServerURL() string {
	return s.Server.URL
}
