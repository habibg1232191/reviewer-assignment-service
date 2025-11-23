// Package dto
package dto

var (
	ErrTeamExist   string = "TEAM_EXIST"
	ErrPrExist     string = "PR_EXIST"
	ErrPrMerged    string = "PR_MERGED"
	ErrNotAssigned string = "NOT_ASSIGNED"
	ErrNoCandidate string = "NO_CANDIDATE"
	ErrNotFound    string = "NOT_FOUND"
	ErrUndefined   string = "UNDEFINED"
)

type ErrorResponseDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorResponseDetail `json:"error"`
}
