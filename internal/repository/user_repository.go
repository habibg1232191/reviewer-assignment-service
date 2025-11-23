package repository

import (
	"context"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
)

type UserRepository interface {
	GetById(ctx context.Context, userID string) (*domain.User, error)
	GetAllByTeamName(ctx context.Context, teamName string) ([]domain.User, error)
	UpdateIsActivity(ctx context.Context, userID string, isActivity bool) (*domain.User, error)
}
