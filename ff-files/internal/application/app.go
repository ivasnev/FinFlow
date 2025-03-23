package application

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-common/logger"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/handler"
	"github.com/ivasnev/FinFlow/ff-files/internal/service/minio"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
	"os"
	"strconv"
)

// App - структура, представляющая приложение с его зависимостями.
type App struct {
	Log       logger.FFLogger
	Config    *config.Config
	Services  *handler.Services
	Handlers  *handler.Handlers
	Router    *gin.Engine
	tvmClient *client.TVMClient
}

// NewApp - конструктор для создания экземпляра приложения с его зависимостями.
func NewApp() (*App, error) {
	// Инициализация логгера
	log := logger.NewLogger("ff-files.log", "debug")
	log.Info("Сервис ff-files запущен!")

	// Загружаем конфигурацию
	appConfig := config.Load()

	// Инициализация клиента Minio
	minioClient := minio.NewMinioService()
	err := minioClient.InitMinio(&appConfig.MinIO)
	if err != nil {
		log.Fatal("Ошибка инициализации Minio: " + err.Error())
		return nil, err
	}

	// Создаем TVM клиент
	tvmClient := client.NewTVMClient(appConfig.TVM.BaseURL, appConfig.TVM.ServiceSecret)

	// Создаем TVM middleware
	tvmMiddleware := middleware.NewTVMMiddleware(tvmClient)

	// Инициализация обработчиков
	minioService, minioHandler := handler.NewHandler(&minioClient)

	// Инициализация маршрутизатора Gin
	router := gin.Default()
	router.Use(gin.RecoveryWithWriter(os.Stdout))

	// Добавляем TVM middleware
	router.Use(tvmMiddleware.ValidateTicket())

	// Создание и возвращение экземпляра приложения
	return &App{
		Log:       log,
		Config:    appConfig,
		Services:  minioService,
		Handlers:  minioHandler,
		Router:    router,
		tvmClient: tvmClient,
	}, nil
}

// Run - метод для запуска приложения.
func (app *App) Run() error {
	// Регистрация маршрутов
	app.Handlers.RegisterRoutes(app.Router)

	// Запуск сервера Gin
	port := strconv.Itoa(app.Config.Server.Port) // Порт берется из конфигурации
	return app.Router.Run(":" + port)
}
