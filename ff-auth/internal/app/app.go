package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
	Config interface{}
}

func New(router *gin.Engine, config interface{}) *App {
	return &App{
		Router: router,
		Config: config,
	}
}

func (a *App) Run(addr string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
	}

	// Запуск сервера в горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("⛔️ Error starting server: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("🚀 Server started on %s\n", addr)

	// Канал для получения сигналов ОС
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокировка до получения сигнала
	<-quit
	fmt.Println("🛑 Shutting down server...")

	// Создаем контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Завершаем сервер
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("⛔️ Server forced to shutdown: %v\n", err)
		return err
	}

	fmt.Println("👋 Server exited")
	return nil
}
