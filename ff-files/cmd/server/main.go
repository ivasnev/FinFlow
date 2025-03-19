package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/client"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ivasnev/FinFlow/ff-files/internal/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/handler"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName)

	// Открываем соединение для миграций
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	// Проверяем подключение к базе данных
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Запускаем миграции
	if err := runMigrations(sqlDB, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Подключаемся через GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Создаем сервис
	svc, err := service.NewService(db, cfg)
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	// Создаем обработчики
	h := handler.NewHandler(svc)

	// Создаем роутер
	r := gin.Default()

	// Создаем TVM клиент
	tvmClient := client.NewTVMClient(cfg.TVM.BaseURL)

	// Создаем TVM middleware
	tvmMiddleware := middleware.NewTVMMiddleware(tvmClient)

	// Добавляем TVM middleware
	r.Use(tvmMiddleware.ValidateTicket())

	// Регистрируем маршруты
	r.POST("/upload", h.UploadFile)
	r.GET("/files/:file_id", h.GetFile)
	r.DELETE("/files/:file_id", h.DeleteFile)
	r.GET("/files/:file_id/metadata", h.GetFileMetadata)
	r.POST("/files/:file_id/temporary-url", h.GenerateTemporaryURL)

	// Запускаем сервер
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
