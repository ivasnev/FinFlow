package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/handlers"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
	_ "github.com/lib/pq"
)

func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
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
			log.Println("Can't close migration:", err.Error())
			return
		}
	}(m)

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к базе данных
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Запуск миграций
	if err := runMigrations(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация компонентов
	repo := service.NewRepository(db)
	keyManager := service.NewKeyManager()
	accessManager := service.NewAccessManager(db)
	ticketService := service.NewTicketService(repo, keyManager, accessManager)
	ticketHandlers := handlers.NewHandlers(ticketService, repo, keyManager)

	// Настройка роутера
	r := gin.Default()

	ticketHandlers.RegisterRoutes(r)

	// Запуск сервера
	addr := fmt.Sprintf(":%s", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
