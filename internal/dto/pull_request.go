// Package dto
package dto

import "time"

type PullRequest struct {
	PullRequestID     string    `json:"pull_request_id"`
	PullRequestName   string    `json:"pull_request_name"`
	AuthorID          string    `json:"authot_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt         time.Time `json:"created_at"`
	MergedAt          time.Time `json:"merged_at"`
}

type PullRequestResponse struct {
	PR PullRequest `json:"pr"`
}

type PullRequestReq struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullRequestReassignReq struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type PullRequestReassignResp struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}
