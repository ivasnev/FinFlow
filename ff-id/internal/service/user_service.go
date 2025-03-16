package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ivasnev/FinFlow/ff-id/internal/config"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailTaken       = errors.New("email already taken")
)

type UserService interface {
	Register(ctx context.Context, email, password, firstName, lastName string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uint) error
	SendVerificationCode(ctx context.Context, userID uint, codeType string) error
	VerifyCode(ctx context.Context, userID uint, code, codeType string) error
	UpdateAvatar(ctx context.Context, userID string, avatarID string) error
}

type userService struct {
	repo   repository.UserRepository
	config *config.Config
}

func NewUserService(repo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		repo:   repo,
		config: cfg,
	}
}

func (s *userService) Register(ctx context.Context, email, password, firstName, lastName string) error {
	// Check if user exists
	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return ErrEmailTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Role:      models.RoleUser,
	}

	return s.repo.Create(ctx, user)
}

func (s *userService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Save session
	session := &models.UserSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.config.JWT.RefreshTTL) * time.Minute),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := s.repo.FindSessionByToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if time.Now().After(session.ExpiresAt) {
		return "", "", errors.New("refresh token expired")
	}

	user, err := s.repo.FindByID(ctx, session.UserID)
	if err != nil {
		return "", "", err
	}

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Update session
	if err := s.repo.DeleteSession(ctx, refreshToken); err != nil {
		return "", "", err
	}

	newSession := &models.UserSession{
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.config.JWT.RefreshTTL) * time.Minute),
	}

	if err := s.repo.CreateSession(ctx, newSession); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	return s.repo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *userService) SendVerificationCode(ctx context.Context, userID uint, codeType string) error {
	code := generateVerificationCode()
	verificationCode := &models.VerificationCode{
		UserID:    userID,
		Code:      code,
		Type:      codeType,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	// TODO: Send code via email or SMS

	return s.repo.CreateVerificationCode(ctx, verificationCode)
}

func (s *userService) VerifyCode(ctx context.Context, userID uint, code, codeType string) error {
	verificationCode, err := s.repo.FindVerificationCode(ctx, userID, codeType)
	if err != nil {
		return err
	}

	if time.Now().After(verificationCode.ExpiresAt) {
		return errors.New("verification code expired")
	}

	if verificationCode.Code != code {
		return errors.New("invalid verification code")
	}

	return nil
}

func (s *userService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(s.config.JWT.AccessTTL) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.AccessSecret))
}

func (s *userService) generateRefreshToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	return token.SignedString([]byte(s.config.JWT.RefreshSecret))
}

func generateVerificationCode() string {
	return rand.New(rand.NewSource(time.Now().UnixNano())).
		Int31n(900000 + 100000).
		String()
}

// UpdateAvatar обновляет ID аватара пользователя
func (s *userService) UpdateAvatar(ctx context.Context, userID string, avatarID string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.AvatarID = avatarID
	return s.repo.Update(ctx, user)
} 