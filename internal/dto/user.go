// Package dto
package dto

type User struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserGetPRReq struct {
	UserID      string             `json:"user_id"`
	PullRequest []PullRequestShort `json:"pull_request"`
}

type UserSetIsActiveReq struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}
