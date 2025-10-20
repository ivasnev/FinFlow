package activity

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// ActivityService реализует интерфейс service.Activity
type ActivityService struct {
	repo repository.Activity
}

// NewActivityService создает новый экземпляр ActivityService
func NewActivityService(repo repository.Activity) *ActivityService {
	return &ActivityService{
		repo: repo,
	}
}

// GetActivitiesByEventID получает активности по ID мероприятия
func (s *ActivityService) GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error) {
	return s.repo.GetByEventID(ctx, eventID)
}

// GetActivityByID получает активность по ID
func (s *ActivityService) GetActivityByID(ctx context.Context, id int) (*models.Activity, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateActivity создает новую активность
func (s *ActivityService) CreateActivity(ctx context.Context, activity *models.Activity) (*models.Activity, error) {
	return s.repo.Create(ctx, activity)
}

// UpdateActivity обновляет активность
func (s *ActivityService) UpdateActivity(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error) {
	return s.repo.Update(ctx, id, activity)
}

// DeleteActivity удаляет активность
func (s *ActivityService) DeleteActivity(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
