package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres"
)

// IconService реализует сервис для работы с иконками
type IconService struct {
	repo *postgres.IconRepository
}

// NewIconService создает новый сервис для работы с иконками
func NewIconService(repo *postgres.IconRepository) *IconService {
	return &IconService{repo: repo}
}

// GetIcons возвращает список всех иконок
func (s *IconService) GetIcons(ctx context.Context) ([]dto.IconFullDTO, error) {
	icons, err := s.repo.GetIcons()
	if err != nil {
		return nil, err
	}

	iconDTOs := make([]dto.IconFullDTO, len(icons))
	for i, icon := range icons {
		iconDTOs[i] = mapIconToDTO(icon)
	}

	return iconDTOs, nil
}

// GetIconByID возвращает иконку по ID
func (s *IconService) GetIconByID(ctx context.Context, id uint) (*dto.IconFullDTO, error) {
	icon, err := s.repo.GetIconByID(id)
	if err != nil {
		return nil, err
	}

	iconDTO := mapIconToDTO(*icon)
	return &iconDTO, nil
}

// CreateIcon создает новую иконку
func (s *IconService) CreateIcon(ctx context.Context, iconDTO *dto.IconFullDTO) (*dto.IconFullDTO, error) {
	icon := mapDTOToIcon(*iconDTO)

	err := s.repo.CreateIcon(&icon)
	if err != nil {
		return nil, err
	}

	result := mapIconToDTO(icon)
	return &result, nil
}

// UpdateIcon обновляет существующую иконку
func (s *IconService) UpdateIcon(ctx context.Context, id uint, iconDTO *dto.IconFullDTO) (*dto.IconFullDTO, error) {
	icon := mapDTOToIcon(*iconDTO)
	icon.ID = id

	err := s.repo.UpdateIcon(&icon)
	if err != nil {
		return nil, err
	}

	result := mapIconToDTO(icon)
	return &result, nil
}

// DeleteIcon удаляет иконку по ID
func (s *IconService) DeleteIcon(ctx context.Context, id uint) error {
	return s.repo.DeleteIcon(id)
}

// Вспомогательные функции для маппинга между моделью и DTO

func mapIconToDTO(icon postgres.Icon) dto.IconFullDTO {
	return dto.IconFullDTO{
		ID:       icon.ID,
		Name:     icon.Name,
		FileUUID: icon.FileUUID,
	}
}

func mapDTOToIcon(iconDTO dto.IconFullDTO) postgres.Icon {
	return postgres.Icon{
		ID:       iconDTO.ID,
		Name:     iconDTO.Name,
		FileUUID: iconDTO.FileUUID,
	}
}
