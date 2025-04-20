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
	"github.com/ivasnev/FinFlow/ff-id/pkg/client"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
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

	// Клиенты внешних сервисов
	AuthClient *auth.Client
	TVMClient  *tvmclient.TVMClient
	IDClient   *client.Client
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

	// Инициализируем ID клиент
	container.IDClient = client.NewClient(
		cfg.IDService.BaseURL,
		cfg.TVM.ServiceID,
		cfg.IDService.ServiceID,
		container.TVMClient,
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

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	// API версии v1
	v1 := c.Router.Group("/api/v1")

	// Middleware для авторизации
	// authMiddleware := auth.AuthMiddleware(c.AuthClient)

	// Middleware для TVM
	tvmMiddleware := tvmmiddleware.NewTVMMiddleware(c.TVMClient)

	// Базовый маршрут для проверки работоспособности сервиса
	v1.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "ok",
			"name":   "FinFlow Split Service",
		})
	})

	// Внутренние маршруты для межсервисного взаимодействия
	internal := c.Router.Group("/internal")
	{
		// Защищенные TVM маршруты
		internalRoutes := internal.Group("/split", tvmMiddleware.ValidateTicket())
		{
			// Примеры маршрутов
			internalRoutes.GET("/health", func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{
					"status": "ok",
					"name":   "FinFlow Split Service Internal",
				})
			})
		}
	}
}
