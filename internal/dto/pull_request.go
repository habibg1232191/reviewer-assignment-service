// Package dto
package dto

import "time"

type PullRequest struct {
	PullRequestID     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"authot_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt         time.Time `json:"create_at"`
	MergedAt          time.Time `json:"merged_at"`
}
