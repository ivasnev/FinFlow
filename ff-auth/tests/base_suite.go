package tests

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/container"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/tests/mockserver"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// BaseSuite представляет базовый набор для интеграционных тестов
type BaseSuite struct {
	suite.Suite
	DBContainer *TestDBContainer
	MockServer  *mockserver.MockServer
	HTTPClient  *http.Client
	Server      *httptest.Server
	Container   *container.Container
	Config      *config.Config
	IDClient    *ffid.Adapter
	APIClient   api.ClientWithResponsesInterface
}

// SetupSuite выполняется один раз перед всеми тестами в suite
func (s *BaseSuite) SetupSuite() {
	// Настройка тестовой БД
	s.DBContainer = setupTestDB(s.T())

	// Настройка мок-сервера
	s.MockServer = setupMockServer(s.T())

	// Создаем конфигурацию для тестов
	s.Config = &config.Config{}
	s.Config.Auth.PasswordHashCost = 10
	s.Config.Auth.AccessTokenDuration = 15
	s.Config.Auth.RefreshTokenDuration = 10080
	s.Config.IDClient.BaseURL = s.MockServer.GetBaseURL()

	// Создаем HTTP клиент без TVM транспорта для тестов
	s.HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Создаем адаптер для ff-id с использованием мок-сервера
	var err error
	s.IDClient, err = ffid.NewAdapter(s.Config.IDClient.BaseURL, s.HTTPClient)
	require.NoError(s.T(), err, "не удалось создать ID клиент")

	// Создаем Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Создаем контейнер с тестовой БД и мок-клиентом
	s.Container, err = createTestContainer(s.T(), s.Config, router, s.DBContainer.DB, s.IDClient)
	require.NoError(s.T(), err, "не удалось создать тестовый контейнер")

	// Регистрируем роуты
	s.Container.RegisterRoutes()

	// Создаем тестовый HTTP сервер
	s.Server = httptest.NewServer(router)

	// Создаем API клиент для удобных запросов
	apiClient, err := api.NewClientWithResponses(s.Server.URL+"/api/v1", api.WithHTTPClient(s.HTTPClient))
	require.NoError(s.T(), err, "не удалось создать API клиент")
	s.APIClient = apiClient
}

// TearDownSuite выполняется один раз после всех тестов в suite
func (s *BaseSuite) TearDownSuite() {
	if s.Server != nil {
		s.Server.Close()
	}
	if s.MockServer != nil {
		s.MockServer.Stop()
	}
	if s.DBContainer != nil {
		teardownTestDB(s.T(), s.DBContainer)
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
		// Сначала очищаем таблицы
		s.DBContainer.DB.Exec("TRUNCATE TABLE devices CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE login_history CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE sessions CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE user_roles CASCADE")
		s.DBContainer.DB.Exec("TRUNCATE TABLE users CASCADE")
		// Затем сбрасываем последовательности
		s.DBContainer.DB.Exec("ALTER SEQUENCE devices_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE login_history_id_seq RESTART WITH 1")
		s.DBContainer.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
		// Не очищаем key_pairs, так как они нужны для TokenManager
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

// GetMockServer возвращает мок-сервер
func (s *BaseSuite) GetMockServer() *mockserver.MockServer {
	return s.MockServer
}
