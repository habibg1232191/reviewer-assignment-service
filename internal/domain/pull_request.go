// Package domain
package domain

import (
	"time"

	"github.com/habibg1232191/reviewer-assignment-service/internal/dto"
)

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    string
	CreatedAt time.Time
	MergedAt  time.Time
}

type PullRequestWithAssignmentReviewers struct {
	PR                  PullRequest
	AssignmentReviewers []string
}

func (prw *PullRequestWithAssignmentReviewers) ToDTO() dto.PullRequest {
	return dto.PullRequest{
		PullRequestID:     prw.PR.ID,
		PullRequestName:   prw.PR.Name,
		AuthorID:          prw.PR.AuthorID,
		Status:            prw.PR.Status,
		AssignedReviewers: prw.AssignmentReviewers,
		CreatedAt:         prw.PR.CreatedAt,
		MergedAt:          prw.PR.MergedAt,
	}
}
