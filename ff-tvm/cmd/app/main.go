package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/handler"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/models"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/repository"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
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
	err = db.AutoMigrate(&models.Service{}, &models.ServiceAccess{}, &models.ServiceTicket{}, &models.KeyRotation{})
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
	serviceRepo := repository.NewServiceRepository(db)
	tvmService := service.NewTVMService(serviceRepo, rdb, cfg)
	tvmHandler := handler.NewTVMHandler(tvmService)

	// Initialize Gin router
	router := gin.Default()

	// TVM routes
	tvm := router.Group("/tvm")
	{
		// Service management
		tvm.POST("/register", tvmHandler.RegisterService)
		tvm.POST("/access/grant", tvmHandler.GrantAccess)
		tvm.POST("/access/revoke", tvmHandler.RevokeAccess)

		// Ticket management
		tvm.POST("/ticket", tvmHandler.IssueTicket)
		tvm.POST("/validate", tvmHandler.ValidateTicket)

		// Key management
		tvm.GET("/public-key/:service_id", tvmHandler.GetPublicKey)
		tvm.POST("/rotate-keys/:service_id", tvmHandler.RotateKeys)
	}

	// Start server
	if err := router.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 