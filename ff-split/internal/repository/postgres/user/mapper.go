package user

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель пользователя БД в бизнес-модель
func extract(dbUser *User) *models.User {
	if dbUser == nil {
		return nil
	}

	return &models.User{
		ID:              dbUser.ID,
		UserID:          dbUser.UserID,
		NicknameCashed:  dbUser.NicknameCashed,
		NameCashed:      dbUser.NameCashed,
		PhotoUUIDCashed: dbUser.PhotoUUIDCashed,
		IsDummy:         dbUser.IsDummy,
	}
}

// extractSlice преобразует слайс моделей пользователей БД в бизнес-модели
func extractSlice(dbUsers []User) []models.User {
	users := make([]models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		if extracted := extract(&dbUser); extracted != nil {
			users[i] = *extracted
		}
	}
	return users
}

// load преобразует бизнес-модель пользователя в модель БД
func load(user *models.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:              user.ID,
		UserID:          user.UserID,
		NicknameCashed:  user.NicknameCashed,
		NameCashed:      user.NameCashed,
		PhotoUUIDCashed: user.PhotoUUIDCashed,
		IsDummy:         user.IsDummy,
	}
}

// extractUserEvent преобразует модель связи БД в бизнес-модель
func extractUserEvent(dbUserEvent *UserEvent) *models.UserEvent {
	if dbUserEvent == nil {
		return nil
	}

	return &models.UserEvent{
		UserID:  dbUserEvent.UserID,
		EventID: dbUserEvent.EventID,
	}
}

// loadUserEvent преобразует бизнес-модель связи в модель БД
func loadUserEvent(userEvent *models.UserEvent) *UserEvent {
	if userEvent == nil {
		return nil
	}

	return &UserEvent{
		UserID:  userEvent.UserID,
		EventID: userEvent.EventID,
	}
}
