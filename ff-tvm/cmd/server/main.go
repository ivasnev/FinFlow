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

func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}
	defer driver.Close()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			logger.ErrorWithStack("Can't close migration", err)
			return
		}
	}(m)

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	logger.Info("Migrations completed successfully")
	return nil
}

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к базе данных
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.FatalWithStack("Failed to connect to database", err)
	}

	// Проверка подключения к базе данных
	if err := db.Ping(); err != nil {
		logger.FatalWithStack("Failed to ping database", err)
	}

	// Запуск миграций
	if err := runMigrations(db, "migrations"); err != nil {
		logger.FatalWithStack("Failed to run migrations", err)
	}
	// Подключение к базе данных
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.FatalWithStack("Failed to connect to database", err)
	}

	// Проверка подключения к базе данных
	if err := db.Ping(); err != nil {
		logger.FatalWithStack("Failed to ping database", err)
	}

	// Инициализация компонентов
	repo := service.NewRepository(db)
	keyManager := service.NewKeyManager()
	accessManager := service.NewAccessManager(db)
	ticketService := service.NewTicketService(repo, keyManager, accessManager)
	ticketHandlers := handlers.NewHandlers(ticketService, repo, keyManager)
	devHandlers := handlers.NewDevHandlers(ticketService)

	// Настройка роутера
	r := gin.Default()

	ticketHandlers.RegisterRoutes(r)
	devHandlers.RegisterRoutes(r)

	// Создаем канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		if err := r.Run(addr); err != nil {
			logger.ErrorWithStack("Failed to start server", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-sigChan
	logger.Info("Shutting down server...")

	if err := db.Close(); err != nil {
		logger.ErrorWithStack("Error closing database connection", err)
	}

	logger.Info("Server stopped")

	// Синхронизируем буфер логгера перед выходом
	if err := logger.Sync(); err != nil {
		fmt.Printf("Error syncing logger: %v\n", err)
	}
}
