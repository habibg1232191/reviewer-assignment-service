// Package postgres
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
	pkgerrors "github.com/habibg1232191/reviewer-assignment-service/pgk/errors"
	"github.com/lib/pq"
)

type ReviewerPostgresRepo struct {
	db *sql.DB
}

func NewPostgresReviewerRepository(db *sql.DB) *ReviewerPostgresRepo {
	return &ReviewerPostgresRepo{db: db}
}

func (r *ReviewerPostgresRepo) CreateTeam(ctx context.Context, team *domain.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	errHandleRollback := func(err error) error {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v; original error: %w", rbErr, err)
				slog.Error("rollback failed", "method", "CreateTeam", "team_name", team.Name, "error", err)
			} else {
				slog.Error("failed to create team", "method", "CreateTeam", "team_name", team.Name, "error", err)
			}

			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				err = pkgerrors.ErrTeamExist
				slog.Error("team_name already exists", "method", "Create", "team_name", team.Name, "error", err)
			}

			return fmt.Errorf("failed to create team: %w", err)
		}
		return nil
	}

	query := "INSERT INTO team(team_name) VALUES($1) RETURNING team_name"
	var teamName string
	err = tx.QueryRowContext(ctx, query, team.Name).Scan(&teamName)
	err = errHandleRollback(err)
	if err != nil {
		return err
	}

	for _, user := range team.Members {
		_, err = tx.ExecContext(
			ctx,
			"INSERT INTO users(user_id, username, team_name, is_active) VALUES($1, $2, $3, $4)",
			user.UserID,
			user.UserName,
			team.Name,
			user.IsActive,
		)
		err = errHandleRollback(err)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

func (r *ReviewerPostgresRepo) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	sqlQuery := `
		SELECT 
			users.user_id,
			users.username,
			users.team_name,
			users.is_active
		FROM users
		JOIN team ON users.team_name = team.team_name
		WHERE team.team_name = $1;
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, teamName)
	if err != nil {
		err := pkgerrors.ErrNotFound
		slog.Error("failed to get team", "method", "GetTeam", "team_name", teamName, "error", err)
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	users := make([]domain.TeamMember, 0)
	for rows.Next() {
		var user domain.User

		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.TeamName,
			&user.IsActive,
		)
		if err != nil {
			slog.Error("failed scan users", "method", "GetTeam", "team_name", teamName, "error", err)
			return nil, fmt.Errorf("failed to get team: %w", err)
		}
		users = append(users, domain.TeamMember{
			UserID:   user.ID,
			UserName: user.Name,
			IsActive: user.IsActive,
		})
	}

	team := &domain.Team{
		Name:    teamName,
		Members: users,
	}

	return team, nil
}

func (r *ReviewerPostgresRepo) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	query := `
	SELECT *
	FROM users
	WHERE user_id=$1
	`

	user := new(domain.User)
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = pkgerrors.ErrNotFound
			slog.Error("not found user", "method", "GetUser", "user_id", userID, "error", err)
			return nil, fmt.Errorf("failed get user: %w", err)
		} else if pqErr, ok := err.(*pq.Error); ok {
			err := fmt.Errorf("pq error: %w; original error: %w", pqErr, err)
			slog.Error("failed get user", "method", "GetUser", "user_id", userID, "error", err)
		} else {
			slog.Error("failed get user; unknow error", "method", "GetUser", "user_id", userID, "error", err)
		}
		return nil, err
	}

	return user, nil
}

func (r *ReviewerPostgresRepo) UserSetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	query := `
		UPDATE users
		SET is_active=$1
		WHERE user_id=$2
		RETURNING users.user_id, users.username, users.team_name, users.is_active
	`

	user := new(domain.User)
	err := r.db.QueryRowContext(ctx, query, isActive, userID).Scan(
		&user.ID,
		&user.Name,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = pkgerrors.ErrNotFound
			slog.Error("the operation failed because the user was not foundfailed because user not found", "method", "UserSetIsActive", "user_id", userID, "error", err)
		} else if pqErr, ok := err.(*pq.Error); ok {
			err = fmt.Errorf("postgres err: %w", pqErr)
			slog.Error("internal error", "method", "UserSetIsActive", "user_id", userID, "error", err)
		} else {
			slog.Error("failed update user", "method", "UserSetIsActive", "user_id", userID, "error", err)
		}

		return nil, fmt.Errorf("failed update user: %w", err)
	}

	return user, nil
}

func (r *ReviewerPostgresRepo) UserGetPR(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query := `
	SELECT pull_request.pr_id, pull_request.pr_name, pull_request.author_id, pull_request.status
	FROM pull_request
	JOIN pr_reviewers ON pr_reviewers.reviewer_id=pull_request.pr_id
	WHERE pull_request.author_id=$1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	prs := make([]domain.PullRequestShort, 0)
	for rows.Next() {
		var pr domain.PullRequestShort
		err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, err
		}

		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *ReviewerPostgresRepo) CreatePullRequest(ctx context.Context, pullRequestID, pullRequestName, authorID string) (*domain.PullRequest, error) {
	query := `
		INSERT INTO pull_request (pr_id, pr_name, author_id)
		VALUES ($1, $2, $3)
		RETURNING pr_id, pr_name, author_id, status, created_at, merged_at;
	`

	pr := new(domain.PullRequest)
	err := r.db.QueryRowContext(ctx, query, pullRequestID, pullRequestName, authorID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = pkgerrors.ErrPrExist
			slog.Error("pull request already exists", "method", "CreatePullRequest", "pull_request_id", pullRequestID, "error", err)
		} else {
			err = fmt.Errorf("error when create pull request: %w", err)
			slog.Error("failed on create pull request", "method", "CreatePullRequest", "pull_request_id", pullRequestID, "error", err)
		}
		return nil, err
	}

	return pr, nil
}

func (r *ReviewerPostgresRepo) AssignmentReviewer(ctx context.Context, prID, reviewerID string) error {
	query := `
	INSERT INTO pr_reviewers(pr_id, reviewer_id) VALUES($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, prID, reviewerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			err = pkgerrors.ErrNotAssigned
			slog.Error("assignment reviewer already exists", "method", "AssignmentReviewer", "pull_request_id", prID, "error", err)
		} else {
			err = fmt.Errorf("internal error: %w", err)
			slog.Error("failed assignment reviewer", "method", "AssignmentReviewer", "pull_request_id", prID, "error", err)
		}
		return fmt.Errorf("failed assignment reviewer: %w", err)
	}
	return nil
}

func (r *ReviewerPostgresRepo) GetAssignmentReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `
	SELECT users.user_id, users.username, users.team_name, users.is_active
	FROM users
	JOIN pr_reviewers ON pr_reviewers.reviewer_id = users.user_id
	WHERE pr_reviewers.pr_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]string, 0)

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *ReviewerPostgresRepo) MarkMerged(ctx context.Context, pullRequestID string) (*domain.PullRequest, error) {
	query := `
	UPDATE pull_request
	SET status='MERGED', merged_at=$1
	WHERE pr_id=$2
	RETURNING *
	`

	var pr domain.PullRequest
	err := r.db.QueryRowContext(ctx, query, time.Now(), pullRequestID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = pkgerrors.ErrNotFound
			slog.Error("not found PR", "method", "MarkMerged", "pull_request", pullRequestID, "error", err)
			return nil, err
		}
		slog.Error("failed to update PR", "method", "MarkMerged", "pull_request", pullRequestID, "error", err)
		return nil, err
	}

	return &pr, nil
}

func (r *ReviewerPostgresRepo) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequestWithAssignmentReviewers, string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", err
	}

	errHandleRollback := func(err error) error {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v; original error: %w", rbErr, err)
				slog.Error("rollback failed", "method", "ReassignReviewer", "pr_id", prID, "error", err)
			} else {
				slog.Error("failed reassign reviewer", "method", "ReassignReviewer", "pr_id", prID, "error", err)
			}
			return fmt.Errorf("failed reassign reviewer: %w", err)
		}
		return nil
	}

	var status string
	err = tx.QueryRowContext(ctx, "SELECT status FROM pull_request WHERE pr_id=$1", prID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", pkgerrors.ErrNotFound
		}
		return nil, "", errHandleRollback(err)
	}
	if status == "MERGED" {
		return nil, "", pkgerrors.ErrPrMerged
	}

	var exists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2)",
		prID, oldReviewerID).Scan(&exists)
	if err != nil {
		return nil, "", errHandleRollback(err)
	}
	if !exists {
		return nil, "", pkgerrors.ErrNotAssigned
	}

	var newReviewerID string
	err = tx.QueryRowContext(ctx, `
		SELECT u.user_id
		FROM users u
		JOIN pull_request pr ON u.team_name = pr_author_team(pr.author_id)
		WHERE u.is_active = TRUE AND u.user_id <> $1
		ORDER BY random()
		LIMIT 1
	`, oldReviewerID).Scan(&newReviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", pkgerrors.ErrNoCandidate
		}
		return nil, "", errHandleRollback(err)
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE pr_reviewers SET reviewer_id=$1 WHERE pr_id=$2 AND reviewer_id=$3",
		newReviewerID, prID, oldReviewerID)
	err = errHandleRollback(err)
	if err != nil {
		return nil, "", err
	}

	var pr domain.PullRequest
	err = tx.QueryRowContext(ctx, `
		SELECT pr_id, pr_name, author_id, status, created_at, merged_at
		FROM pull_request WHERE pr_id=$1
	`, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	err = errHandleRollback(err)
	if err != nil {
		return nil, "", err
	}

	rows, err := tx.QueryContext(ctx,
		"SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1", prID)
	if err != nil {
		return nil, "", errHandleRollback(err)
	}
	defer rows.Close()

	prs := make([]string, 0)

	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, "", errHandleRollback(err)
		}
		prs = append(prs, rid)
	}

	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("commit failed: %w", err)
	}

	prsRes := &domain.PullRequestWithAssignmentReviewers{
		PR:                  pr,
		AssignmentReviewers: prs,
	}

	return prsRes, newReviewerID, nil
}
