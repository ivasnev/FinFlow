package friend

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// ExtractUserFriend преобразует модель связи дружбы базы данных в доменную модель
func ExtractUserFriend(dbFriend *UserFriend) *models.UserFriend {
	if dbFriend == nil {
		return nil
	}

	return &models.UserFriend{
		ID:        dbFriend.ID,
		UserID:    dbFriend.UserID,
		FriendID:  dbFriend.FriendID,
		Status:    dbFriend.Status,
		CreatedAt: dbFriend.CreatedAt,
	}
}

// LoadUserFriend преобразует доменную модель связи дружбы в модель базы данных
func LoadUserFriend(friend *models.UserFriend) *UserFriend {
	if friend == nil {
		return nil
	}

	return &UserFriend{
		ID:        friend.ID,
		UserID:    friend.UserID,
		FriendID:  friend.FriendID,
		Status:    friend.Status,
		CreatedAt: friend.CreatedAt,
	}
}

// ExtractUser преобразует модель пользователя базы данных в доменную модель
func ExtractUser(dbUser *User) models.User {
	var name sql.NullString
	if dbUser.Name != "" {
		name.String = dbUser.Name
		name.Valid = true
	}

	var avatarID uuid.NullUUID
	if dbUser.AvatarID != "" {
		parsedUUID, err := uuid.Parse(dbUser.AvatarID)
		if err == nil {
			avatarID.UUID = parsedUUID
			avatarID.Valid = true
		}
	}

	return models.User{
		ID:       dbUser.ID,
		Email:    dbUser.Email,
		Nickname: dbUser.Nickname,
		Name:     name,
		AvatarID: avatarID,
	}
}
