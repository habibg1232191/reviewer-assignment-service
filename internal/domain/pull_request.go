// Package domain
package domain

import "time"

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          time.Time
}
