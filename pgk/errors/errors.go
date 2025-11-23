// Package errors
package errors

import "errors"

var (
	ErrTeamExist   error = errors.New("TEAM_EXIST")
	ErrPrExist     error = errors.New("PR_EXIST")
	ErrPrMerged    error = errors.New("PR_MERGED")
	ErrNotAssigned error = errors.New("NOT_ASSIGNED")
	ErrNoCandidate error = errors.New("NO_CANDIDATE")
	ErrNotFound    error = errors.New("NOT_FOUND")
)
