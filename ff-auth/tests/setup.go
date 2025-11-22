package tests

import (
	"context"
	_ "embed"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/tests/mockserver"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations.sql
var migrationSQL string

// TestDBContainer представляет контейнер с тестовой БД
type TestDBContainer struct {
	Container testcontainers.Container
	DB        *gorm.DB
}

// setupTestDB создает тестовую базу данных с использованием testcontainers
func setupTestDB(t *testing.T) *TestDBContainer {
	ctx := context.Background()

	// Создаем PostgreSQL контейнер
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("ff_auth_test"),
		postgrescontainer.WithUsername("postgres"),
		postgrescontainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "не удалось запустить PostgreSQL контейнер")

	// Получаем строку подключения
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "не удалось получить строку подключения")

	// Подключаемся к БД
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err, "не удалось подключиться к тестовой БД")

	// Выполняем SQL миграции из встроенного файла
	err = db.Exec(migrationSQL).Error
	require.NoError(t, err, "не удалось выполнить миграции")

	// Создаем роль "user" если её нет (используем прямое SQL, так как таблица уже создана миграцией)
	var roleID int
	result := db.Raw("SELECT id FROM roles WHERE name = ?", string(models.RoleUser)).Scan(&roleID)
	if result.Error != nil || roleID == 0 {
		// Роль уже должна быть создана миграцией, но на всякий случай проверяем
		err = db.Exec("INSERT INTO roles (name) VALUES (?) ON CONFLICT (name) DO NOTHING", string(models.RoleUser)).Error
		require.NoError(t, err, "не удалось создать роль user")
	}

	return &TestDBContainer{
		Container: postgresContainer,
		DB:        db,
	}
}

// teardownTestDB останавливает и удаляет тестовый контейнер
func teardownTestDB(t *testing.T, container *TestDBContainer) {
	ctx := context.Background()
	if container != nil && container.Container != nil {
		// Очищаем данные из таблиц перед остановкой контейнера
		if container.DB != nil {
			container.DB.Exec("TRUNCATE TABLE devices CASCADE")
			container.DB.Exec("TRUNCATE TABLE login_history CASCADE")
			container.DB.Exec("TRUNCATE TABLE sessions CASCADE")
			container.DB.Exec("TRUNCATE TABLE user_roles CASCADE")
			container.DB.Exec("TRUNCATE TABLE roles CASCADE")
			container.DB.Exec("TRUNCATE TABLE users CASCADE")
			container.DB.Exec("TRUNCATE TABLE key_pairs CASCADE")
		}
		// Останавливаем контейнер
		err := container.Container.Terminate(ctx)
		require.NoError(t, err, "не удалось остановить контейнер")
	}
}

// setupMockServer создает и запускает мок-сервер
func setupMockServer(t *testing.T) *mockserver.MockServer {
	// httptest.Server запускается автоматически
	server := mockserver.NewMockServer()
	return server
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestMain выполняется один раз перед всеми тестами
func TestMain(m *testing.M) {
	// Устанавливаем режим Gin в test
	gin.SetMode(gin.TestMode)

	// Запускаем тесты
	code := m.Run()

	// Выходим с кодом возврата
	os.Exit(code)
}
