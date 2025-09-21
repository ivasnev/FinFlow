package main

import (
	"fmt"
	"log"

	"github.com/ivasnev/FinFlow/ff-id/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/container"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация роутера Gin
	router := gin.Default()

	// Инициализация контейнера зависимостей
	c, err := container.NewContainer(cfg, router)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Регистрация маршрутов
	c.RegisterRoutes()

	// Создание и запуск приложения
	application := app.New(router, cfg)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err := application.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
