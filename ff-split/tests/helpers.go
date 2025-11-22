package tests

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// createTestUser создает тестового пользователя в БД
func (s *BaseSuite) createTestUser(id int64, userID int64, nickname, name string) *models.User {
	userIDPtr := &userID
	user := &models.User{
		ID:             id,
		UserID:         userIDPtr,
		NicknameCashed: nickname,
		NameCashed:     name,
		IsDummy:        false,
	}

	// Создаем пользователя напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO users (id, user_id, nickname_cashed, name_cashed, is_dummy)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, *user.UserID, user.NicknameCashed, user.NameCashed, user.IsDummy).Error
	s.NoError(err, "не удалось создать тестового пользователя")

	return user
}

// createTestDummyUser создает dummy-пользователя в БД
func (s *BaseSuite) createTestDummyUser(id int64, nickname, name string) *models.User {
	user := &models.User{
		ID:             id,
		UserID:         nil, // Dummy users don't have external user_id
		NicknameCashed: nickname,
		NameCashed:     name,
		IsDummy:        true,
	}

	// Создаем пользователя напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO users (id, nickname_cashed, name_cashed, is_dummy)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.NicknameCashed, user.NameCashed, user.IsDummy).Error
	s.NoError(err, "не удалось создать dummy-пользователя")

	return user
}

// createTestIcon создает тестовую иконку в БД
func (s *BaseSuite) createTestIcon(id int, name, fileUUID string) *models.Icon {
	icon := &models.Icon{
		ID:       id,
		Name:     name,
		FileUUID: fileUUID,
	}

	// Создаем иконку напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO icons (id, name, file_uuid)
		VALUES ($1, $2, $3)
	`, icon.ID, icon.Name, icon.FileUUID).Error
	s.NoError(err, "не удалось создать тестовую иконку")

	return icon
}

// createTestEventCategory создает тестовую категорию мероприятия в БД
func (s *BaseSuite) createTestEventCategory(id int, name string, iconID int) *models.EventCategory {
	category := &models.EventCategory{
		ID:     id,
		Name:   name,
		IconID: iconID,
	}

	// Создаем категорию напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO event_categories (id, name, icon_id)
		VALUES ($1, $2, $3)
	`, category.ID, category.Name, category.IconID).Error
	s.NoError(err, "не удалось создать тестовую категорию мероприятия")

	return category
}

// createTestTransactionCategory создает тестовую категорию транзакции в БД
func (s *BaseSuite) createTestTransactionCategory(id int, name string, iconID int) *models.TransactionCategory {
	category := &models.TransactionCategory{
		ID:     id,
		Name:   name,
		IconID: iconID,
	}

	// Создаем категорию напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO transaction_categories (id, name, icon_id)
		VALUES ($1, $2, $3)
	`, category.ID, category.Name, category.IconID).Error
	s.NoError(err, "не удалось создать тестовую категорию транзакции")

	return category
}

// createTestEvent создает тестовое мероприятие в БД
func (s *BaseSuite) createTestEvent(id int64, name, description string, categoryID *int) *models.Event {
	event := &models.Event{
		ID:          id,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		Status:      "active",
	}

	// Создаем мероприятие напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO events (id, name, description, category_id, status)
		VALUES ($1, $2, $3, $4, $5)
	`, event.ID, event.Name, event.Description, event.CategoryID, event.Status).Error
	s.NoError(err, "не удалось создать тестовое мероприятие")

	return event
}

// addUserToEvent добавляет пользователя к мероприятию
func (s *BaseSuite) addUserToEvent(userID int64, eventID int64) {
	err := s.GetDB().Exec(`
		INSERT INTO user_event (user_id, event_id)
		VALUES ($1, $2)
	`, userID, eventID).Error
	s.NoError(err, "не удалось добавить пользователя к мероприятию")
}

