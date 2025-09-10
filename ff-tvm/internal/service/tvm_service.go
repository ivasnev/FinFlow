package service

import (
	"context"
	"encoding/json"
	"errors"
	"ff-tvm/internal/config"
	"ff-tvm/internal/models"
	"ff-tvm/internal/repository"
	"ff-tvm/pkg/crypto"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrServiceExists   = errors.New("service already exists")
	ErrAccessDenied    = errors.New("access denied")
	ErrInvalidTicket   = errors.New("invalid ticket")
	ErrTicketExpired   = errors.New("ticket expired")
)

type TVMService interface {
	RegisterService(ctx context.Context, name, description string) (*models.Service, error)
	GrantAccess(ctx context.Context, sourceID, targetID uint) error
	RevokeAccess(ctx context.Context, sourceID, targetID uint) error
	IssueTicket(ctx context.Context, sourceID, targetID uint) (string, error)
	ValidateTicket(ctx context.Context, ticketStr string) (map[string]interface{}, error)
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
	if _, err := s.repo.GetServiceByName(ctx, name); err == nil {
		return nil, ErrServiceExists
	}

	publicKey, privateKey, err := crypto.GenerateEd25519KeyPair()
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

type TicketClaims struct {
	SourceID  uint  `json:"source_id"`
	TargetID  uint  `json:"target_id"`
	IssuedAt  int64 `json:"issued_at"`
	ExpiresAt int64 `json:"expires_at"`
}

func (s *tvmService) IssueTicket(ctx context.Context, sourceID, targetID uint) (string, error) {
	hasAccess, err := s.repo.CheckServiceAccess(ctx, sourceID, targetID)
	if err != nil {
		return "", err
	}
	if !hasAccess {
		return "", ErrAccessDenied
	}

	source, err := s.repo.GetServiceByID(ctx, sourceID)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.ParseEd25519PrivateKey(source.PrivateKey)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(s.config.TVM.TicketTTL)
	claims := TicketClaims{
		SourceID:  sourceID,
		TargetID:  targetID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signature, err := crypto.SignEd25519(privateKey, claimsJSON)
	if err != nil {
		return "", err
	}

	ticket := struct {
		Claims    string `json:"claims"`
		Signature string `json:"signature"`
	}{
		Claims:    string(claimsJSON),
		Signature: signature,
	}

	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		return "", err
	}

	// Сохраняем тикет
	ticketModel := &models.ServiceTicket{
		SourceServiceID: sourceID,
		TargetServiceID: targetID,
		Token:           string(ticketJSON),
		ExpiresAt:       expiresAt,
	}

	if err := s.repo.CreateServiceTicket(ctx, ticketModel); err != nil {
		return "", err
	}

	// Кешируем тикет в Redis
	key := s.ticketCacheKey(sourceID, targetID)
	if err := s.redis.Set(ctx, key, string(ticketJSON), s.config.TVM.TicketTTL).Err(); err != nil {
		// Логируем ошибку, но продолжаем работу
		// TODO: добавить логирование
	}

	return string(ticketJSON), nil
}

func (s *tvmService) ValidateTicket(ctx context.Context, ticketStr string) (map[string]interface{}, error) {
	var ticket struct {
		Claims    string `json:"claims"`
		Signature string `json:"signature"`
	}

	if err := json.Unmarshal([]byte(ticketStr), &ticket); err != nil {
		return nil, ErrInvalidTicket
	}

	var claims TicketClaims
	if err := json.Unmarshal([]byte(ticket.Claims), &claims); err != nil {
		return nil, ErrInvalidTicket
	}

	// Проверяем срок действия
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTicketExpired
	}

	// Получаем сервис-источник для его публичного ключа
	source, err := s.repo.GetServiceByID(ctx, claims.SourceID)
	if err != nil {
		return nil, err
	}

	publicKey, err := crypto.ParseEd25519PublicKey(source.PublicKey)
	if err != nil {
		return nil, err
	}

	// Проверяем подпись
	if err := crypto.VerifyEd25519(publicKey, []byte(ticket.Claims), ticket.Signature); err != nil {
		return nil, ErrInvalidTicket
	}

	// Проверяем права доступа
	hasAccess, err := s.repo.CheckServiceAccess(ctx, claims.SourceID, claims.TargetID)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrAccessDenied
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(ticket.Claims), &result); err != nil {
		return nil, err
	}

	return result, nil
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

	newPublicKey, newPrivateKey, err := crypto.GenerateEd25519KeyPair()
	if err != nil {
		return err
	}

	rotation := &models.KeyRotation{
		ServiceID: serviceID,
		OldKey:    service.PublicKey,
		NewKey:    newPublicKey,
	}

	if err := s.repo.CreateKeyRotation(ctx, rotation); err != nil {
		return err
	}

	service.PublicKey = newPublicKey
	service.PrivateKey = newPrivateKey

	return s.repo.UpdateService(ctx, service)
}

func (s *tvmService) ticketCacheKey(sourceID, targetID uint) string {
	return "ticket:" + string(sourceID) + ":" + string(targetID)
}
