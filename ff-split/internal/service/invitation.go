package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Invitation управляет отзывными ссылками и явным вступлением по ним.
type Invitation interface {
	List(context.Context, int64) ([]models.Invitation, error)
	Create(context.Context, int64) (*models.Invitation, string, error)
	Revoke(context.Context, int64, int64) error
	Preview(context.Context, string) (*models.InvitationPreview, error)
	Accept(context.Context, string) (int64, error)
}
