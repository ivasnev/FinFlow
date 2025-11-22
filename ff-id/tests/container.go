package tests

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/container"
	avatarRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/avatar"
	friendRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/friend"
	userRepo "github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres/user"
	friendService "github.com/ivasnev/FinFlow/ff-id/internal/service/friend"
	userService "github.com/ivasnev/FinFlow/ff-id/internal/service/user"
	"gorm.io/gorm"
)

// createTestContainer создает тестовый контейнер с роутером для HTTP сервера
func createTestContainer(t *testing.T, cfg *config.Config, router *gin.Engine, db *gorm.DB) (*container.Container, error) {
	c := &container.Container{
		Config: cfg,
		Router: router,
		DB:     db,
	}

	// Инициализируем репозитории
	c.UserRepository = userRepo.NewUserRepository(c.DB)
	c.AvatarRepository = avatarRepo.NewAvatarRepository(c.DB)
	c.FriendRepository = friendRepo.NewFriendRepository(c.DB)

	// Инициализируем сервисы
	c.UserService = userService.NewUserService(c.UserRepository, c.AvatarRepository)
	c.FriendService = friendService.NewFriendService(c.FriendRepository, c.UserRepository)

	// Инициализируем обработчики (без клиентов для тестов)
	// Обработчики будут созданы в базовом suite

	return c, nil
}
