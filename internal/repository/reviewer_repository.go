// Package repository
package repository

import (
	"context"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
)

type ReviewerRepository interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	UserGetPR(ctx context.Context, userID string) ([]domain.PullRequestShort, error)

	CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string) (*domain.PullRequest, error)
	AssignmentReviewer(ctx context.Context, prID, reviewerID string) error
	GetAssignmentReviewers(ctx context.Context, prID string) ([]string, error)
	MarkMerged(ctx context.Context, pullRequestID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequestWithAssignmentReviewers, string, error)
}
