// Package usecase
package usecase

import (
	"context"
	"reviewer-assignment-service/internal/domain"
	"reviewer-assignment-service/internal/repository"
)

type ReviewerSerivice interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type reviewerService struct {
	reviewerRepo repository.ReviewerRepository
}

func NewReviwerService(reviewerRepo repository.ReviewerRepository) *reviewerService {
	return &reviewerService{reviewerRepo: reviewerRepo}
}

func (r *reviewerService) CreateTeam(ctx context.Context, team *domain.Team) error {
	return r.reviewerRepo.CreateTeam(ctx, team)
}

func (r *reviewerService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return r.reviewerRepo.GetTeam(ctx, teamName)
}

func (r *reviewerService) UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return r.reviewerRepo.UserSetIsActive(ctx, userID, isActive)
}
