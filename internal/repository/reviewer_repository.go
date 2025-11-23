// Package repository
package repository

import (
	"context"
	"reviewer-assignment-service/internal/domain"
)

type ReviewerRepository interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}
