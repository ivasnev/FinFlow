package login_history

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// ExtractLoginHistory преобразует модель истории входов базы данных в обычную модель
func ExtractLoginHistory(dbHistory *LoginHistory) *models.LoginHistory {
	if dbHistory == nil {
		return nil
	}

	return &models.LoginHistory{
		ID:        dbHistory.ID,
		UserID:    dbHistory.UserID,
		IPAddress: dbHistory.IPAddress,
		UserAgent: dbHistory.UserAgent,
		CreatedAt: dbHistory.CreatedAt,
	}
}

// loadLoginHistory преобразует обычную модель истории входов в модель базы данных
func loadLoginHistory(history *models.LoginHistory) *LoginHistory {
	if history == nil {
		return nil
	}

	return &LoginHistory{
		ID:        history.ID,
		UserID:    history.UserID,
		IPAddress: history.IPAddress,
		UserAgent: history.UserAgent,
		CreatedAt: history.CreatedAt,
	}
}
