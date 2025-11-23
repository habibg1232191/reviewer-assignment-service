package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/habibg1232191/reviewer-assignment-service/internal/domain"
	pkgerrors "github.com/habibg1232191/reviewer-assignment-service/pgk/errors"
	"github.com/lib/pq"
)

type UserPostgresRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{db: db}
}

func (u *UserPostgresRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE user_id=$1
	`

	user := new(domain.User)
	err := u.db.QueryRowContext(ctx, query, userID).Scan(user)
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
