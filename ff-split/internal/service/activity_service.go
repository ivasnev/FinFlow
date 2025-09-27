package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// ActivityServiceImpl реализация сервиса активностей
type ActivityServiceImpl struct {
	activityRepo repository.ActivityRepository
}

// NewActivityService создает новый экземпляр сервиса активностей
func NewActivityService(activityRepo repository.ActivityRepository) *ActivityServiceImpl {
	return &ActivityServiceImpl{
		activityRepo: activityRepo,
	}
}

// GetActivitiesByEventID возвращает все активности по ID мероприятия
func (s *ActivityServiceImpl) GetActivitiesByEventID(ctx context.Context, eventID int64) ([]dto.ActivityResponse, error) {
	activities, err := s.activityRepo.GetActivitiesByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	var response []dto.ActivityResponse
	for _, activity := range activities {
		response = append(response, mapActivityToResponse(activity))
	}

	return response, nil
}

// GetActivityByID возвращает активность по ID
func (s *ActivityServiceImpl) GetActivityByID(ctx context.Context, id int) (dto.ActivityResponse, error) {
	activity, err := s.activityRepo.GetActivityByID(ctx, id)
	if err != nil {
		return dto.ActivityResponse{}, err
	}

	return mapActivityToResponse(activity), nil
}

// CreateActivity создает новую активность
func (s *ActivityServiceImpl) CreateActivity(ctx context.Context, activity models.Activity) (dto.ActivityResponse, error) {
	// Установка текущего времени если не указано
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}

	createdActivity, err := s.activityRepo.CreateActivity(ctx, activity)
	if err != nil {
		return dto.ActivityResponse{}, err
	}

	return mapActivityToResponse(createdActivity), nil
}

// UpdateActivity обновляет существующую активность
func (s *ActivityServiceImpl) UpdateActivity(ctx context.Context, activity models.Activity) error {
	return s.activityRepo.UpdateActivity(ctx, activity)
}

// DeleteActivity удаляет активность
func (s *ActivityServiceImpl) DeleteActivity(ctx context.Context, id int) error {
	return s.activityRepo.DeleteActivity(ctx, id)
}

// mapActivityToResponse преобразует модель Activity в DTO
func mapActivityToResponse(activity models.Activity) dto.ActivityResponse {
	return dto.ActivityResponse{
		ActivityID:  activity.ID,
		Description: activity.Description,
		IconID:      "", // Поле для иконки, может быть заполнено позже
		DateTime:    activity.CreatedAt,
	}
}
