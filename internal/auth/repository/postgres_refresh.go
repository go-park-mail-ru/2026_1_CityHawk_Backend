package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
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

func (r *PostgresRefreshRepository) Store(refreshToken, userID string, expiresAt time.Time) error {
	_, err := r.pool.Exec(
		context.Background(),
		storeRefreshSessionQuery,
		postgresTokenHash(refreshToken),
		userID,
		expiresAt,
	)
	return err
}

func (r *PostgresRefreshRepository) Consume(refreshToken string) (string, error) {
	var userID string
	var expiresAt time.Time
	err := r.pool.QueryRow(
		context.Background(),
		consumeRefreshSessionQuery,
		postgresTokenHash(refreshToken),
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

func (r *PostgresRefreshRepository) Revoke(refreshToken string) error {
	_, err := r.pool.Exec(
		context.Background(),
		revokeRefreshSessionQuery,
		postgresTokenHash(refreshToken),
	)
	return err
}

func postgresTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
