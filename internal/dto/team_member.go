// Package dto
package dto

type TeamMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	IsActive bool   `json:"is_active"`
}
