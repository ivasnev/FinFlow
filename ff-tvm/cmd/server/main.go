package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/handlers"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/logger"
	_ "github.com/lib/pq"
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

func runMigrations(cfg *config.Config, log *logger.Logger) error {
	db, err := connectToDatabase(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", cfg.Migrations.Path),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			log.ErrorWithStack("Can't close migration", err)
			return
		}
	}(m)

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Info("Migrations completed successfully")
	return nil
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

	// Запуск миграций
	if err := runMigrations(cfg, log); err != nil {
		log.FatalWithStack("Failed to run migrations", err)
	}

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
	ticketHandlers := handlers.NewHandlers(ticketService, repo, keyManager)

	// Инициализация обработчиков для разработчиков только если включен режим разработки
	var devHandlers *handlers.DevHandlers
	if cfg.Dev.Enabled {
		devHandlers = handlers.NewDevHandlers(ticketService, cfg)
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
