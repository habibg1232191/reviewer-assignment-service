// Package postgres
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reviewer-assignment-service/internal/domain"

	"github.com/lib/pq"
)

type ReviewerPostgresRepo struct {
	db *sql.DB
}

func NewPostgresReviewerRepository(db *sql.DB) *ReviewerPostgresRepo {
	return &ReviewerPostgresRepo{db: db}
}

func (r *ReviewerPostgresRepo) CreateTeam(ctx context.Context, team domain.Team) error {
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

	tx.Commit()
	return nil
}

func (r *ReviewerPostgresRepo) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	sqlQuery := `
		SELECT * FROM users
		JOIN team ON users.team_name=team.team_name
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery)
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

	var user *domain.User = new(domain.User)
	err := r.db.QueryRowContext(ctx, query, userID).Scan(user)
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
		RETURNING *
	`

	var user *domain.User = new(domain.User)
	err := r.db.QueryRowContext(ctx, query, isActive, userID).Scan(user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = pkgerrors.ErrNotFound
			slog.Error("the operation failed because the user was not foundfailed because user not found", "method", "UserSetIsActive", "user_id", userID, "error", err)
		} else if pgErr, ok := err.(*pq.Error); ok {
			err = fmt.Errorf("postgres err: %w", pgErr)
			slog.Error("internal error", "method", "UserSetIsActive", "user_id", userID, "error", err)
		} else {
			slog.Error("failed update user", "method", "UserSetIsActive", "user_id", userID, "error", err)
		}

		return nil, fmt.Errorf("failed update user: %w", err)
	}

	return user, nil
}
