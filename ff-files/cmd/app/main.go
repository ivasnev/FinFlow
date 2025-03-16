package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	tvmclient "github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	tvmmiddleware "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
	"github.com/ivasnev/FinFlow/ff-files/internal/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/handler"
	"github.com/ivasnev/FinFlow/ff-files/internal/repository"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	"github.com/ivasnev/FinFlow/ff-files/internal/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создание хранилища файлов
	localStorage, err := storage.NewLocalStorage(cfg.Storage.BasePath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// Создание TVM клиента
	tvmClient := tvmclient.NewClient(tvmclient.Config{
		BaseURL:    cfg.TVM.BaseURL,
		ServiceID:  cfg.TVM.ServiceID,
		ServiceKey: cfg.TVM.ServiceKey,
	})

	// Создание репозитория
	fileRepo := repository.NewFileRepository(db)

	// Создание сервиса
	fileService := service.NewFileService(fileRepo, localStorage, &service.Config{
		MaxFileSize:       cfg.Storage.MaxFileSize,
		AllowedMimeTypes: cfg.Storage.AllowedMimeTypes,
		SoftDeleteTimeout: cfg.Storage.SoftDeleteTimeout,
	})

	// Создание обработчика
	fileHandler := handler.NewFileHandler(fileService)

	// Настройка роутера
	router := gin.Default()

	// Добавляем middleware для проверки TVM тикетов
	router.Use(tvmmiddleware.TVMAuth(tvmClient))

	// Регистрируем маршруты
	fileHandler.RegisterRoutes(router)

	// Создание HTTP-сервера
	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: router,
	}

	// Запуск сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Запуск периодической очистки в горутине
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := fileService.CleanupExpiredFiles(context.Background()); err != nil {
					log.Printf("Failed to cleanup expired files: %v", err)
				}
				if err := fileService.CleanupExpiredURLs(context.Background()); err != nil {
					log.Printf("Failed to cleanup expired URLs: %v", err)
				}
			}
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
} 