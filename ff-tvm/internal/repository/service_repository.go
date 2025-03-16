package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/models"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	CreateService(ctx context.Context, service *models.Service) error
	GetServiceByID(ctx context.Context, id uint) (*models.Service, error)
	GetServiceByName(ctx context.Context, name string) (*models.Service, error)
	UpdateService(ctx context.Context, service *models.Service) error
	DeleteService(ctx context.Context, id uint) error
	
	CreateServiceAccess(ctx context.Context, access *models.ServiceAccess) error
	CheckServiceAccess(ctx context.Context, sourceID, targetID uint) (bool, error)
	DeleteServiceAccess(ctx context.Context, sourceID, targetID uint) error
	
	CreateServiceTicket(ctx context.Context, ticket *models.ServiceTicket) error
	GetServiceTicket(ctx context.Context, sourceID, targetID uint) (*models.ServiceTicket, error)
	
	CreateKeyRotation(ctx context.Context, rotation *models.KeyRotation) error
	GetLastKeyRotation(ctx context.Context, serviceID uint) (*models.KeyRotation, error)
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) CreateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *serviceRepository) GetServiceByID(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).First(&service, id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) GetServiceByName(ctx context.Context, name string) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) UpdateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Save(service).Error
}

func (r *serviceRepository) DeleteService(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Service{}, id).Error
}

func (r *serviceRepository) CreateServiceAccess(ctx context.Context, access *models.ServiceAccess) error {
	return r.db.WithContext(ctx).Create(access).Error
}

func (r *serviceRepository) CheckServiceAccess(ctx context.Context, sourceID, targetID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ServiceAccess{}).
		Where("source_service_id = ? AND target_service_id = ? AND (expires_at IS NULL OR expires_at > NOW())",
			sourceID, targetID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *serviceRepository) DeleteServiceAccess(ctx context.Context, sourceID, targetID uint) error {
	return r.db.WithContext(ctx).
		Where("source_service_id = ? AND target_service_id = ?", sourceID, targetID).
		Delete(&models.ServiceAccess{}).Error
}

func (r *serviceRepository) CreateServiceTicket(ctx context.Context, ticket *models.ServiceTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *serviceRepository) GetServiceTicket(ctx context.Context, sourceID, targetID uint) (*models.ServiceTicket, error) {
	var ticket models.ServiceTicket
	err := r.db.WithContext(ctx).
		Where("source_service_id = ? AND target_service_id = ? AND expires_at > NOW()",
			sourceID, targetID).
		First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *serviceRepository) CreateKeyRotation(ctx context.Context, rotation *models.KeyRotation) error {
	return r.db.WithContext(ctx).Create(rotation).Error
}

func (r *serviceRepository) GetLastKeyRotation(ctx context.Context, serviceID uint) (*models.KeyRotation, error) {
	var rotation models.KeyRotation
	err := r.db.WithContext(ctx).
		Where("service_id = ?", serviceID).
		Order("created_at DESC").
		First(&rotation).Error
	if err != nil {
		return nil, err
	}
	return &rotation, nil
} 