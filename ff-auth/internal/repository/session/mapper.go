package session

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// ExtractSession преобразует модель сессии базы данных в обычную модель
func ExtractSession(dbSession *Session) *models.Session {
	if dbSession == nil {
		return nil
	}

	return &models.Session{
		ID:           dbSession.ID,
		UserID:       dbSession.UserID,
		RefreshToken: dbSession.RefreshToken,
		IPAddress:    dbSession.IPAddress,
		ExpiresAt:    dbSession.ExpiresAt,
		CreatedAt:    dbSession.CreatedAt,
	}
}

// loadSession преобразует обычную модель сессии в модель базы данных
func loadSession(session *models.Session) *Session {
	if session == nil {
		return nil
	}

	return &Session{
		ID:           session.ID,
		UserID:       session.UserID,
		RefreshToken: session.RefreshToken,
		IPAddress:    session.IPAddress,
		ExpiresAt:    session.ExpiresAt,
		CreatedAt:    session.CreatedAt,
	}
}
