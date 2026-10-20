package container

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-files/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-files/internal/api/middleware"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	minioService "github.com/ivasnev/FinFlow/ff-files/internal/service/minio"
	"github.com/ivasnev/FinFlow/ff-files/pkg/api"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmmiddleware "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
)

// Container - контейнер зависимостей для приложения
type Container struct {
	Config        *config.Config
	Router        *gin.Engine
	MinIOService  service.MinIO
	ServerHandler *handler.ServerHandler
	TVMClient     tvmmiddleware.TVMClient
}

// NewContainer - конструктор контейнера зависимостей
func NewContainer(cfg *config.Config, router *gin.Engine) (*Container, error) {
	container := &Container{
		Config:    cfg,
		Router:    router,
		TVMClient: tvmclient.NewTVMClient(cfg.TVM.BaseURL, cfg.TVM.ServiceSecret),
	}

	// Инициализируем MinIO сервис
	if err := container.initMinIO(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации MinIO: %w", err)
	}

	// Инициализируем обработчики
	container.initHandlers()

	return container, nil
}

// initMinIO инициализирует MinIO сервис
func (c *Container) initMinIO() error {
	minioService := minioService.NewMinioService()

	// Инициализируем подключение к MinIO
	if err := minioService.InitMinio(&c.Config.MinIO); err != nil {
		return fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	c.MinIOService = minioService
	return nil
}

// initHandlers инициализирует обработчики
func (c *Container) initHandlers() {
	c.ServerHandler = handler.NewServerHandler(c.MinIOService)
}

// RegisterRoutes - регистрирует все маршруты API
func (c *Container) RegisterRoutes() {
	c.Router.Use(middleware.CORSMiddleware())
	// API версии v1
	ticketMiddleware := tvmmiddleware.NewTVMMiddleware(c.TVMClient, c.Config.TVM.ServiceID)
	v1 := c.Router.Group("", ticketMiddleware.ValidateTicket())

	// Регистрируем маршруты с помощью сгенерированного сервера
	api.RegisterHandlersWithOptions(v1, c.ServerHandler, api.GinServerOptions{})
}
