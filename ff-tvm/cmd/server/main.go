package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	handlers2 "github.com/ivasnev/FinFlow/ff-tvm/internal/api/handlers"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/logger"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
	"syscall"
)

func connectToDatabase(cfg *config.Config, log *logger.Logger) (*sql.DB, error) {
	// Подключение к базе данных
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.FatalWithStack("Failed to connect to database", err)
		return nil, err
	}

	// Проверка подключения к базе данных
	if err := db.Ping(); err != nil {
		log.FatalWithStack("Failed to ping database", err)
		return nil, err
	}

	return db, nil
}

func main() {
	// Инициализация логгера
	log, err := logger.New()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к базе данных
	db, err := connectToDatabase(cfg, log)
	if err != nil {
		log.FatalWithStack("Failed to connect to database", err)
	}
	defer db.Close()

	// Инициализация компонентов
	repo := service.NewRepository(db)
	keyManager := service.NewKeyManager()
	accessManager := service.NewAccessManager(db)
	ticketService := service.NewTicketService(repo, keyManager, accessManager)
	ticketHandlers := handlers2.NewHandlers(ticketService, repo, keyManager)

	// Инициализация обработчиков для разработчиков только если включен режим разработки
	var devHandlers *handlers2.DevHandlers
	if cfg.Dev.Enabled {
		devHandlers = handlers2.NewDevHandlers(ticketService, cfg)
		log.Info("Development mode enabled")
	}

	// Настройка роутера
	r := gin.Default()

	ticketHandlers.RegisterRoutes(r)
	if cfg.Dev.Enabled {
		devHandlers.RegisterRoutes(r)
	}

	// Создаем канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		if err := r.Run(addr); err != nil {
			log.ErrorWithStack("Failed to start server", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-sigChan
	log.Info("Shutting down server...")

	if err := db.Close(); err != nil {
		log.ErrorWithStack("Error closing database connection", err)
	}

	log.Info("Server stopped")
}
