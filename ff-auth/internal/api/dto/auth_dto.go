package dto

import "time"

// RegisterRequest представляет запрос на регистрацию нового пользователя
type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone,omitempty" binding:"omitempty,e164"`
	Password string  `json:"password" binding:"required,min=8"`
	Nickname string  `json:"nickname" binding:"required,min=3,max=50"`
	Name     *string `json:"name,omitempty"`
}

// LoginRequest представляет запрос на вход в систему
type LoginRequest struct {
	Login    string `json:"login" binding:"required"` // Может быть email или nickname
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest представляет запрос на обновление access-токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse представляет ответ после успешной аутентификации
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         ShortUserDTO `json:"user"`
}

// LogoutRequest представляет запрос на выход из системы
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
