package avatar

import (
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// ExtractUserAvatar преобразует модель аватара базы данных в доменную модель
func ExtractUserAvatar(dbAvatar *UserAvatar) *models.UserAvatar {
	if dbAvatar == nil {
		return nil
	}

	return &models.UserAvatar{
		ID:         dbAvatar.ID,
		UserID:     dbAvatar.UserID,
		FileID:     dbAvatar.FileID,
		UploadedAt: dbAvatar.UploadedAt,
	}
}

// LoadUserAvatar преобразует доменную модель аватара в модель базы данных
func LoadUserAvatar(avatar *models.UserAvatar) *UserAvatar {
	if avatar == nil {
		return nil
	}

	return &UserAvatar{
		ID:         avatar.ID,
		UserID:     avatar.UserID,
		FileID:     avatar.FileID,
		UploadedAt: avatar.UploadedAt,
	}
}
