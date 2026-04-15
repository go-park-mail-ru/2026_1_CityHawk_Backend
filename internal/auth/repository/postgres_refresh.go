package repository

import (
	"context"
	"errors"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
	platformsecurity "cityhawk/backend/internal/platform/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRefreshRepository struct {
	pool *pgxpool.Pool
}

const (
	storeRefreshSessionQuery = `
		INSERT INTO refresh_session (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_hash) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    expires_at = EXCLUDED.expires_at
	`
	consumeRefreshSessionQuery = `
		DELETE FROM refresh_session
		WHERE token_hash = $1
		RETURNING user_id, expires_at
	`
	revokeRefreshSessionQuery = `
		DELETE FROM refresh_session
		WHERE token_hash = $1
	`
)

func NewPostgresRefreshRepository(pool *pgxpool.Pool) *PostgresRefreshRepository {
	return &PostgresRefreshRepository{pool: pool}
}

func (r *PostgresRefreshRepository) Store(ctx context.Context, refreshToken, userID string, expiresAt time.Time) error {
	_, err := r.pool.Exec(
		ctx,
		storeRefreshSessionQuery,
		platformsecurity.SHA256Hex(refreshToken),
		userID,
		expiresAt,
	)
	return err
}

func (r *PostgresRefreshRepository) Consume(ctx context.Context, refreshToken string) (string, error) {
	var userID string
	var expiresAt time.Time
	err := r.pool.QueryRow(
		ctx,
		consumeRefreshSessionQuery,
		platformsecurity.SHA256Hex(refreshToken),
	).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", platformerrors.ErrTokenRevoked
		}
		return "", platformerrors.ErrInternal
	}

	if time.Now().UTC().After(expiresAt) {
		return "", platformerrors.ErrTokenExpired
	}

	return userID, nil
}

func (r *PostgresRefreshRepository) Revoke(ctx context.Context, refreshToken string) error {
	_, err := r.pool.Exec(
		ctx,
		revokeRefreshSessionQuery,
		platformsecurity.SHA256Hex(refreshToken),
	)
	return err
}
