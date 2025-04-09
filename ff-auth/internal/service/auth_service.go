package service

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/postgres"
	idclient "github.com/ivasnev/FinFlow/ff-id/pkg/client"
	"golang.org/x/crypto/bcrypt"
)

// AuthService реализует интерфейс для аутентификации и авторизации
type AuthService struct {
	config                 *config.Config
	userRepository         postgres.UserRepositoryInterface
	roleRepository         postgres.RoleRepositoryInterface
	sessionRepository      postgres.SessionRepositoryInterface
	deviceService          DeviceServiceInterface
	loginHistoryRepository postgres.LoginHistoryRepositoryInterface
	tokenManager           *ED25519TokenManager
	idClient               *idclient.Client
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(
	config *config.Config,
	userRepository postgres.UserRepositoryInterface,
	roleRepository postgres.RoleRepositoryInterface,
	sessionRepository postgres.SessionRepositoryInterface,
	deviceService DeviceServiceInterface,
	loginHistoryRepository postgres.LoginHistoryRepositoryInterface,
	tokenManager *ED25519TokenManager,
	idClient *idclient.Client,
) *AuthService {
	return &AuthService{
		config:                 config,
		userRepository:         userRepository,
		roleRepository:         roleRepository,
		sessionRepository:      sessionRepository,
		deviceService:          deviceService,
		loginHistoryRepository: loginHistoryRepository,
		tokenManager:           tokenManager,
		idClient:               idClient,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Проверяем, существует ли пользователь с таким никнеймом
	existingUser, err = s.userRepository.GetByNickname(ctx, req.Nickname)
	if err == nil && existingUser != nil {
		return nil, errors.New("пользователь с таким никнеймом уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.Auth.PasswordHashCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// Создаем нового пользователя
	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Сохраняем пользователя в базе данных
	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	reqRegister := &idclient.RegisterUserRequest{
		Email:    user.Email,
		Nickname: user.Nickname,
		UserID:   user.ID,
	}

	_, err = s.idClient.RegisterUser(ctx, reqRegister)
	if err != nil {
		dbErr := s.userRepository.Delete(ctx, user.ID)
		if dbErr != nil {
			return nil, fmt.Errorf("ошибка удаления пользователя из базы данных: %w", dbErr)
		}
		return nil, fmt.Errorf("ошибка регистрации пользователя в ID: %w", err)
	}

	// Назначаем пользователю роль "user"
	userRole, err := s.roleRepository.GetByName(ctx, string(models.RoleUser))
	if err != nil {
		return nil, fmt.Errorf("ошибка получения роли: %w", err)
	}

	if err := s.userRepository.AddRole(ctx, user.ID, userRole.ID); err != nil {
		return nil, fmt.Errorf("ошибка назначения роли: %w", err)
	}

	// Создаем пару токенов для пользователя
	accessToken, refreshToken, expiresAt, err := s.GenerateTokenPair(ctx, user.ID, []string{string(models.RoleUser)})
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	// Создаем сессию
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepository.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	// Формируем ответ
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		User: dto.ShortUserDTO{
			ID:       user.ID,
			Email:    user.Email,
			Nickname: user.Nickname,
			Roles:    []string{string(models.RoleUser)},
		},
	}, nil
}

// Login выполняет вход пользователя в систему
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, r *http.Request) (*dto.AuthResponse, error) {
	var user *models.User
	var err error

	// Пытаемся найти пользователя по email или никнейму
	if strings.Contains(req.Login, "@") {
		user, err = s.userRepository.GetByEmail(ctx, req.Login)
	} else {
		user, err = s.userRepository.GetByNickname(ctx, req.Login)
	}

	if err != nil {
		return nil, errors.New("неверный логин или пароль")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("неверный логин или пароль")
	}

	// Получаем роли пользователя
	roles, err := s.userRepository.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ролей пользователя: %w", err)
	}

	// Преобразуем роли в строки
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Name
	}

	// Создаем пару токенов для пользователя
	accessToken, refreshToken, expiresAt, err := s.GenerateTokenPair(ctx, user.ID, roleStrings)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	// Определяем устройство пользователя
	userAgent := r.UserAgent()
	deviceID := generateDeviceID(r)

	// Получаем или создаем устройство
	_, err = s.deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка работы с устройством: %w", err)
	}

	// Создаем сессию
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepository.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	// Записываем историю входа
	if err := s.RecordLogin(ctx, user.ID, r); err != nil {
		// Не фатальная ошибка, просто логируем
		fmt.Printf("Ошибка записи истории входа: %v\n", err)
	}

	// Формируем DTO для пользователя
	userDTO := dto.ShortUserDTO{
		ID:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
		Roles:    roleStrings,
	}

	// Формируем ответ
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		User:         userDTO,
	}, nil
}

// RefreshToken обновляет access-токен
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// Находим сессию по refresh-токену
	session, err := s.sessionRepository.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("недействительный refresh-токен")
	}

	// Проверяем, не истек ли срок действия токена
	if session.ExpiresAt.Before(time.Now()) {
		// Удаляем истекшую сессию
		_ = s.sessionRepository.Delete(ctx, session.ID)
		return nil, errors.New("истек срок действия refresh-токена")
	}

	// Получаем пользователя
	user, err := s.userRepository.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден: %w", err)
	}

	// Получаем роли пользователя
	roles, err := s.userRepository.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ролей пользователя: %w", err)
	}

	// Преобразуем роли в строки
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Name
	}

	// Создаем новую пару токенов
	accessToken, newRefreshToken, expiresAt, err := s.GenerateTokenPair(ctx, user.ID, roleStrings)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации токенов: %w", err)
	}

	// Обновляем сессию с новым refresh-токеном
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Unix(expiresAt, 0)

	if err := s.sessionRepository.Delete(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("ошибка удаления старой сессии: %w", err)
	}

	// Создаем новую сессию
	newSession := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepository.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("ошибка создания новой сессии: %w", err)
	}

	// Формируем DTO для пользователя
	userDTO := dto.ShortUserDTO{
		ID:       user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
		Roles:    roleStrings,
	}

	// Формируем ответ
	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		User:         userDTO,
	}, nil
}

// Logout выполняет выход пользователя из системы
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Находим сессию по refresh-токену
	session, err := s.sessionRepository.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return errors.New("недействительный refresh-токен")
	}

	// Удаляем сессию
	return s.sessionRepository.Delete(ctx, session.ID)
}

// GenerateTokenPair генерирует пару токенов (access и refresh)
func (s *AuthService) GenerateTokenPair(ctx context.Context, userID int64, roles []string) (accessToken, refreshToken string, expiresAt int64, err error) {
	// Генерируем токены с помощью tokenManager
	accessTTL := time.Duration(s.config.Auth.AccessTokenDuration) * time.Minute
	refreshTTL := time.Duration(s.config.Auth.RefreshTokenDuration) * time.Minute // В минутах как в конфиге

	return s.tokenManager.GenerateTokenPair(userID, roles, accessTTL, refreshTTL)
}

// ValidateToken проверяет валидность токена
func (s *AuthService) ValidateToken(tokenString string) (int64, []string, error) {
	// Проверяем токен с помощью tokenManager
	payload, err := s.tokenManager.ValidateToken(tokenString)
	if err != nil {
		return 0, nil, err
	}

	return payload.UserID, payload.Roles, nil
}

// RecordLogin записывает историю входа
func (s *AuthService) RecordLogin(ctx context.Context, userID int64, r *http.Request) error {
	// Получаем IP-адрес пользователя
	ipAddress := extractIPAddress(r)

	// Создаем запись в истории входов
	loginHistory := &models.LoginHistory{
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: r.UserAgent(),
		CreatedAt: time.Now(),
	}

	return s.loginHistoryRepository.Create(ctx, loginHistory)
}

// GetPublicKey возвращает публичный ключ для проверки токенов
func (s *AuthService) GetPublicKey() ed25519.PublicKey {
	return s.tokenManager.GetPublicKey()
}

// Вспомогательные функции

// extractIPAddress извлекает IP-адрес из запроса
func extractIPAddress(r *http.Request) string {
	// Проверяем заголовок X-Forwarded-For
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Берем первый IP из списка
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Проверяем заголовок X-Real-IP
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Получаем IP из RemoteAddr
	ip := r.RemoteAddr
	// Удаляем порт, если он есть
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	return ip
}

// generateDeviceID генерирует идентификатор устройства из запроса
func generateDeviceID(r *http.Request) string {
	// Генерируем хеш на основе User-Agent и IP-адреса
	return fmt.Sprintf("%s_%s", r.UserAgent(), extractIPAddress(r))
}
