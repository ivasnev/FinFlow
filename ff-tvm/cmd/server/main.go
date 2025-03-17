package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/handlers"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
	_ "github.com/lib/pq"
)

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
