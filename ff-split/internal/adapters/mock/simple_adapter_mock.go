package mock

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/adapters"
)

// SimpleIDAdapter представляет простой мок-адаптер для интеграционных тестов
type SimpleIDAdapter struct {
	users map[int64]*adapters.UserDTO
}

// NewSimpleIDAdapter создает новый простой мок-адаптер
func NewSimpleIDAdapter() *SimpleIDAdapter {
	return &SimpleIDAdapter{
		users: make(map[int64]*adapters.UserDTO),
	}
}

// SetupDefaultBehavior настраивает поведение мока по умолчанию
func (m *SimpleIDAdapter) SetupDefaultBehavior() {
	// Добавляем тестовых пользователей
	name1 := "User One"
	name2 := "User Two"
	name3 := "User Three"

	m.users[1] = &adapters.UserDTO{
		ID:        1,
		Nickname:  "user1",
		Name:      &name1,
		CreatedAt: 1724684557,
		UpdatedAt: 1724684557,
	}
	m.users[2] = &adapters.UserDTO{
		ID:        2,
		Nickname:  "user2",
		Name:      &name2,
		CreatedAt: 1724684557,
		UpdatedAt: 1724684557,
	}
	m.users[3] = &adapters.UserDTO{
		ID:        3,
		Nickname:  "user3",
		Name:      &name3,
		CreatedAt: 1724684557,
		UpdatedAt: 1724684557,
	}
}

// AddUser добавляет пользователя в мок
func (m *SimpleIDAdapter) AddUser(user *adapters.UserDTO) {
	m.users[user.ID] = user
}

// GetUsersByIDs возвращает пользователей по их ID
func (m *SimpleIDAdapter) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]adapters.UserDTO, error) {
	var users []adapters.UserDTO
	for _, id := range userIDs {
		if user, ok := m.users[id]; ok {
			users = append(users, *user)
		}
	}
	return users, nil
}

// GetUserByID возвращает пользователя по ID
func (m *SimpleIDAdapter) GetUserByID(ctx context.Context, userID int64) (*adapters.UserDTO, error) {
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

// Проверяем, что SimpleIDAdapter реализует интерфейс IDAdapter
var _ adapters.IDAdapter = (*SimpleIDAdapter)(nil)

