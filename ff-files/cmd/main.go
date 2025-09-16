package main

import (
	"ff-files/internal/application"
	"go.uber.org/zap"
)

func main() {
	// Создаем экземпляр приложения
	app, err := application.NewApp()
	if err != nil {
		panic(err)
		return
	}

	// Используем логгер с полями
	app.Log.With(zap.String("service", "ff-files")).Info("Старт работы сервиса")

	// Запуск приложения
	if err := app.Run(); err != nil {
		app.Log.Fatal("Ошибка запуска сервера Gin: " + err.Error())
	}
}
