// Package usecase
package usecase

import (
	"context"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
	"github.com/habibg1232191/reviewer-assignment-service/internal/repository"
)

type ReviewerService interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	UserGetPR(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string) (*domain.PullRequestWithAssignmentReviewers, error)
	MarkMerged(ctx context.Context, pullRequestID string) (*domain.PullRequestWithAssignmentReviewers, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequestWithAssignmentReviewers, string, error)
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

func (r *reviewerService) UserGetPR(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	return r.reviewerRepo.UserGetPR(ctx, userID)
}

func (r *reviewerService) CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string) (*domain.PullRequestWithAssignmentReviewers, error) {
	prCreated, err := r.reviewerRepo.CreatePullRequest(ctx, pullRequestID, pullRequestName, authorID)
	if err != nil {
		return nil, err
	}

	user, err := r.reviewerRepo.GetUser(ctx, authorID)
	if err != nil {
		return nil, err
	}

	team, err := r.reviewerRepo.GetTeam(ctx, user.TeamName)
	if err != nil {
		return nil, err
	}

	if len(team.Members) >= 2 {
		count := 0
		for _, member := range team.Members {
			if count > 2 {
				break
			}
			if member.IsActive {
				r.reviewerRepo.AssignmentReviewer(ctx, prCreated.AuthorID, member.UserID)
				count += 1
			}
		}
	} else if len(team.Members) == 1 && team.Members[0].IsActive {
		r.reviewerRepo.AssignmentReviewer(ctx, prCreated.AuthorID, team.Members[0].UserID)
	}

	reviewers, err := r.reviewerRepo.GetAssignmentReviewers(ctx, prCreated.ID)
	if err != nil {
		return nil, err
	}

	prRes := &domain.PullRequestWithAssignmentReviewers{
		PR:                  *prCreated,
		AssignmentReviewers: reviewers,
	}

	return prRes, nil
}

func (r *reviewerService) MarkMerged(ctx context.Context, pullRequestID string) (*domain.PullRequestWithAssignmentReviewers, error) {
	pr, err := r.reviewerRepo.MarkMerged(ctx, pullRequestID)
	if err != nil {
		return nil, err
	}
	reviewers, err := r.reviewerRepo.GetAssignmentReviewers(ctx, pr.ID)
	if err != nil {
		return nil, err
	}

	prRes := &domain.PullRequestWithAssignmentReviewers{
		PR:                  *pr,
		AssignmentReviewers: reviewers,
	}

	return prRes, nil
}

func (r *reviewerService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequestWithAssignmentReviewers, string, error) {
	return r.reviewerRepo.ReassignReviewer(ctx, prID, oldReviewerID)
}
