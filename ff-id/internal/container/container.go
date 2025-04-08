package container

import (
	"fmt"

	"github.com/gin-gonic/gin"
	// "github.com/ivasnev/FinFlow/ff-auth/internal/api/middleware"
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
	UserRepository   pg_repos.UserRepositoryInterface
	AvatarRepository pg_repos.AvatarRepositoryInterface

	// Сервисы
	UserService service.UserServiceInterface

	// Обработчики
	UserHandler *handler.UserHandler
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
	c.AvatarRepository = pg_repos.NewAvatarRepository(c.DB)
}

// initServices инициализирует сервисы
func (c *Container) initServices() {
	c.UserService = service.NewUserService(c.UserRepository, c.AvatarRepository)
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.UserHandler = handler.NewUserHandler(c.UserService)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Middleware для авторизации
	// authMiddleware := middleware.AuthMiddleware(c.AuthService)

	// Группа маршрутов для пользователей
	users := v1.Group("/users")
	{
		// Публичные маршруты
		users.GET("/:nickname", c.UserHandler.GetUserByNickname)

		// Защищенные маршруты
		// users.Use(authMiddleware)
		users.PATCH("/me", c.UserHandler.UpdateUser)
	}
}
