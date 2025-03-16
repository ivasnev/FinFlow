package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/config"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/models"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/repository"
	"github.com/ivasnev/FinFlow/ff-tvm/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrServiceNotFound     = errors.New("service not found")
	ErrServiceExists       = errors.New("service already exists")
	ErrAccessDenied       = errors.New("access denied")
	ErrInvalidTicket      = errors.New("invalid ticket")
	ErrTicketExpired      = errors.New("ticket expired")
)

type TVMService interface {
	RegisterService(ctx context.Context, name, description string) (*models.Service, error)
	GrantAccess(ctx context.Context, sourceID, targetID uint) error
	RevokeAccess(ctx context.Context, sourceID, targetID uint) error
	IssueTicket(ctx context.Context, sourceID, targetID uint) (string, error)
	ValidateTicket(ctx context.Context, ticketStr string) (*jwt.MapClaims, error)
	GetPublicKey(ctx context.Context, serviceID uint) (string, error)
	RotateKeys(ctx context.Context, serviceID uint) error
}

type tvmService struct {
	repo   repository.ServiceRepository
	redis  *redis.Client
	config *config.Config
}

func NewTVMService(repo repository.ServiceRepository, redis *redis.Client, cfg *config.Config) TVMService {
	return &tvmService{
		repo:   repo,
		redis:  redis,
		config: cfg,
	}
}

func (s *tvmService) RegisterService(ctx context.Context, name, description string) (*models.Service, error) {
	// Проверяем, не существует ли уже сервис с таким именем
	if _, err := s.repo.GetServiceByName(ctx, name); err == nil {
		return nil, ErrServiceExists
	}

	// Генерируем пару ключей RSA
	publicKey, privateKey, err := crypto.GenerateKeyPair(s.config.TVM.RSAKeyBits)
	if err != nil {
		return nil, err
	}

	service := &models.Service{
		Name:        name,
		Description: description,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Active:      true,
	}

	if err := s.repo.CreateService(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *tvmService) GrantAccess(ctx context.Context, sourceID, targetID uint) error {
	// Проверяем существование сервисов
	if _, err := s.repo.GetServiceByID(ctx, sourceID); err != nil {
		return ErrServiceNotFound
	}
	if _, err := s.repo.GetServiceByID(ctx, targetID); err != nil {
		return ErrServiceNotFound
	}

	access := &models.ServiceAccess{
		SourceServiceID: sourceID,
		TargetServiceID: targetID,
	}

	return s.repo.CreateServiceAccess(ctx, access)
}

func (s *tvmService) RevokeAccess(ctx context.Context, sourceID, targetID uint) error {
	return s.repo.DeleteServiceAccess(ctx, sourceID, targetID)
}

func (s *tvmService) IssueTicket(ctx context.Context, sourceID, targetID uint) (string, error) {
	// Проверяем права доступа
	hasAccess, err := s.repo.CheckServiceAccess(ctx, sourceID, targetID)
	if err != nil {
		return "", err
	}
	if !hasAccess {
		return "", ErrAccessDenied
	}

	// Получаем сервис-источник для его приватного ключа
	source, err := s.repo.GetServiceByID(ctx, sourceID)
	if err != nil {
		return "", err
	}

	// Парсим приватный ключ
	privateKey, err := crypto.ParsePrivateKey(source.PrivateKey)
	if err != nil {
		return "", err
	}

	// Создаем JWT токен
	expiresAt := time.Now().Add(s.config.TVM.TicketTTL)
	claims := jwt.MapClaims{
		"source_id":  sourceID,
		"target_id":  targetID,
		"exp":        expiresAt.Unix(),
		"issued_at":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	// Сохраняем тикет
	ticket := &models.ServiceTicket{
		SourceServiceID: sourceID,
		TargetServiceID: targetID,
		Token:          tokenString,
		ExpiresAt:      expiresAt,
	}

	if err := s.repo.CreateServiceTicket(ctx, ticket); err != nil {
		return "", err
	}

	// Кешируем тикет в Redis
	key := s.ticketCacheKey(sourceID, targetID)
	if err := s.redis.Set(ctx, key, tokenString, s.config.TVM.TicketTTL).Err(); err != nil {
		// Логируем ошибку, но продолжаем работу
		// TODO: добавить логирование
	}

	return tokenString, nil
}

func (s *tvmService) ValidateTicket(ctx context.Context, ticketStr string) (*jwt.MapClaims, error) {
	// Парсим токен без проверки подписи
	token, _ := jwt.Parse(ticketStr, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidTicket
	}

	sourceID := uint(claims["source_id"].(float64))
	targetID := uint(claims["target_id"].(float64))

	// Получаем сервис-источник для его публичного ключа
	source, err := s.repo.GetServiceByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// Парсим публичный ключ
	publicKey, err := crypto.ParsePublicKey(source.PublicKey)
	if err != nil {
		return nil, err
	}

	// Проверяем подпись и валидность токена
	token, err = jwt.Parse(ticketStr, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidTicket
	}

	// Проверяем права доступа
	hasAccess, err := s.repo.CheckServiceAccess(ctx, sourceID, targetID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrAccessDenied
	}

	return &claims, nil
}

func (s *tvmService) GetPublicKey(ctx context.Context, serviceID uint) (string, error) {
	service, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return "", ErrServiceNotFound
	}
	return service.PublicKey, nil
}

func (s *tvmService) RotateKeys(ctx context.Context, serviceID uint) error {
	service, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return ErrServiceNotFound
	}

	// Генерируем новую пару ключей
	newPublicKey, newPrivateKey, err := crypto.GenerateKeyPair(s.config.TVM.RSAKeyBits)
	if err != nil {
		return err
	}

	// Сохраняем старые ключи в истории
	rotation := &models.KeyRotation{
		ServiceID: serviceID,
		OldKey:    service.PublicKey,
		NewKey:    newPublicKey,
	}

	if err := s.repo.CreateKeyRotation(ctx, rotation); err != nil {
		return err
	}

	// Обновляем ключи сервиса
	service.PublicKey = newPublicKey
	service.PrivateKey = newPrivateKey

	return s.repo.UpdateService(ctx, service)
}

func (s *tvmService) ticketCacheKey(sourceID, targetID uint) string {
	return "ticket:" + string(sourceID) + ":" + string(targetID)
} 