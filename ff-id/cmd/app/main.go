package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/handler"
	"github.com/ivasnev/FinFlow/ff-id/internal/middleware"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	dsn := "host=" + cfg.Database.Host +
		" user=" + cfg.Database.User +
		" password=" + cfg.Database.Password +
		" dbname=" + cfg.Database.DBName +
		" port=" + cfg.Database.Port +
		" sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate database schemas
	err = db.AutoMigrate(&models.User{}, &models.UserSession{}, &models.VerificationCode{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize dependencies
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)
	userHandler := handler.NewUserHandler(userService)

	// Initialize Gin router
	router := gin.Default()

	// Public routes
	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)
	router.POST("/refresh", userHandler.RefreshToken)

	// Protected routes
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
	{
		protected.GET("/profile", userHandler.GetProfile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.DELETE("/profile", userHandler.DeleteProfile)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg.JWT.AccessSecret))
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
	{
		// Add admin routes here
	}

	// Start server
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 