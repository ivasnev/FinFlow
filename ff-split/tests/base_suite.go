package tests

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/internal/container"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/ivasnev/FinFlow/ff-split/tests/mockserver"
	testTime "github.com/ivasnev/FinFlow/ff-split/tests/utils/time"
	testUUID "github.com/ivasnev/FinFlow/ff-split/tests/utils/uuid"
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

	// Мок-серверы для внешних зависимостей
	FFIDMockServer *mockserver.MockServer

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

	// 4. Настройка мок-сервера для ff-id
	s.FFIDMockServer = setupMockServer(s.T())

	// 5. Создаем конфигурацию для тестов
	s.Config = &config.Config{}
	s.Config.Postgres.Host = "localhost"
	s.Config.Postgres.Port = 5432
	s.Config.Postgres.User = "postgres"
	s.Config.Postgres.Password = "postgres"
	s.Config.Postgres.DBName = "ff_split_test"
	// Настраиваем ff-id клиент на мок-сервер
	s.Config.IDService.BaseURL = s.FFIDMockServer.GetBaseURL()

	// 6. Создаем HTTP клиент для тестов
	s.HTTPClient = &http.Client{}

	// 7. Создаем Gin router в тестовом режиме
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 8. Создаем контейнер с тестовой БД и HTTP клиентом
	var err error
	s.Container, err = createTestContainer(s.T(), s.Config, router, s.DBContainer.DB, s.HTTPClient)
	require.NoError(s.T(), err, "не удалось создать тестовый контейнер")

	// 9. Создаем обработчики
	s.Container.ServerHandler = handler.NewServerHandler(
		s.Container.EventService,
		s.Container.UserService,
		s.Container.TransactionService,
		s.Container.ActivityService,
		s.Container.TaskService,
		s.Container.CategoryService,
		s.Container.IconService,
	)

	// 10. Тестовый middleware для установки user_id
	testAuthMiddleware := func(c *gin.Context) {
		// Устанавливаем тестовый user_id для всех запросов
		c.Set("user_id", TestUserID1)
		c.Next()
	}

	// 11. Регистрируем роуты с тестовым middleware
	v1 := router.Group("/api/v1")
	v1.Use(testAuthMiddleware)
	api.RegisterHandlers(v1, s.Container.ServerHandler)

	// 12. Создаем тестовый HTTP сервер
	s.Server = httptest.NewServer(router)

	// 13. Создаем API клиент для удобных запросов
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

	// Закрытие мок-сервера
	if s.FFIDMockServer != nil {
		s.FFIDMockServer.Stop()
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
	// Очищаем мок-сервер и проверяем, что все ожидаемые вызовы были сделаны
	if s.FFIDMockServer != nil {
		s.FFIDMockServer.Clear(s.T())
	}

	// Очищаем данные после каждого теста
	s.cleanupDatabase()
}

// cleanupDatabase очищает тестовую базу данных
func (s *BaseSuite) cleanupDatabase() {
	if s.DBContainer != nil && s.DBContainer.DB != nil {
		// Выполняем очистку в правильном порядке из-за внешних ключей
		s.DBContainer.DB.Exec("TRUNCATE TABLE optimized_debts CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE debts CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE transaction_shares CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE transactions CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE tasks CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE activities CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE user_event CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE events CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE transaction_categories CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE event_categories CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE icons CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE users CASCADE")

		// Сбрасываем последовательности
		s.DBContainer.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE events_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE transactions_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE transaction_shares_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE debts_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE tasks_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE activities_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE icons_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE event_categories_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE transaction_categories_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE optimized_debts_id_seq RESTART WITH 1")
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

// GetMockServer возвращает мок-сервер для ff-id
func (s *BaseSuite) GetMockServer() *mockserver.MockServer {
	return s.FFIDMockServer
}
