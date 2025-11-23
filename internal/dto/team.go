// Package dto
package dto

type TeamReq struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamResponse struct {
	Team TeamReq `json:"team"`
}
