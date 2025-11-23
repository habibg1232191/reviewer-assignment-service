// Package dto
package dto

type TeamMember struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	IsActive bool   `json:"is_active"`
}
