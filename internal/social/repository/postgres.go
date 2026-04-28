package repository

import (
	"context"
	"database/sql"
	"errors"

	platformerrors "cityhawk/backend/internal/platform/errors"
	socialmodel "cityhawk/backend/internal/social/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) AddFavorite(ctx context.Context, userID, eventID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO favorite_event (user_id, event_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, eventID)
	return mapPgError(err)
}

func (r *PostgresRepository) RemoveFavorite(ctx context.Context, userID, eventID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM favorite_event WHERE user_id = $1 AND event_id = $2`, userID, eventID)
	return err
}

func (r *PostgresRepository) ListFavoriteEvents(ctx context.Context, userID string, limit, offset int) ([]socialmodel.FavoriteEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, created_at
		FROM favorite_event
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []socialmodel.FavoriteEvent
	for rows.Next() {
		var item socialmodel.FavoriteEvent
		if err := rows.Scan(&item.EventID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) IsFavorite(ctx context.Context, userID, eventID string) (socialmodel.FavoriteFlag, error) {
	var flag socialmodel.FavoriteFlag
	var createdAt sql.NullTime
	err := r.pool.QueryRow(ctx, `
		SELECT event_id, true, created_at
		FROM favorite_event
		WHERE user_id = $1 AND event_id = $2
	`, userID, eventID).Scan(&flag.EventID, &flag.IsFavorite, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return socialmodel.FavoriteFlag{EventID: eventID}, nil
	}
	if err != nil {
		return socialmodel.FavoriteFlag{}, err
	}
	if createdAt.Valid {
		value := createdAt.Time.UTC()
		flag.CreatedAt = &value
	}
	return flag, nil
}

func (r *PostgresRepository) FollowUser(ctx context.Context, followerUserID, followedUserID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_follow (follower_user_id, followed_user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, followerUserID, followedUserID)
	return mapPgError(err)
}

func (r *PostgresRepository) UnfollowUser(ctx context.Context, followerUserID, followedUserID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_follow WHERE follower_user_id = $1 AND followed_user_id = $2`, followerUserID, followedUserID)
	return err
}

func (r *PostgresRepository) ListFollowers(ctx context.Context, userID string, limit, offset int) ([]socialmodel.UserFollow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT follower_user_id, created_at
		FROM user_follow
		WHERE followed_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return scanUserFollows(rows, err)
}

func (r *PostgresRepository) ListFollowing(ctx context.Context, userID string, limit, offset int) ([]socialmodel.UserFollow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT followed_user_id, created_at
		FROM user_follow
		WHERE follower_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	return scanUserFollows(rows, err)
}

func (r *PostgresRepository) IsFollowing(ctx context.Context, followerUserID, followedUserID string) (socialmodel.FollowingFlag, error) {
	var flag socialmodel.FollowingFlag
	var createdAt sql.NullTime
	err := r.pool.QueryRow(ctx, `
		SELECT followed_user_id, true, created_at
		FROM user_follow
		WHERE follower_user_id = $1 AND followed_user_id = $2
	`, followerUserID, followedUserID).Scan(&flag.UserID, &flag.IsFollowing, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return socialmodel.FollowingFlag{UserID: followedUserID}, nil
	}
	if err != nil {
		return socialmodel.FollowingFlag{}, err
	}
	if createdAt.Valid {
		value := createdAt.Time.UTC()
		flag.CreatedAt = &value
	}
	return flag, nil
}

func (r *PostgresRepository) FavoriteFlags(ctx context.Context, userID string, eventIDs []string) (map[string]socialmodel.FavoriteFlag, error) {
	flags := make(map[string]socialmodel.FavoriteFlag, len(eventIDs))
	for _, id := range eventIDs {
		flags[id] = socialmodel.FavoriteFlag{EventID: id}
	}
	if userID == "" {
		return flags, nil
	}

	for _, eventID := range eventIDs {
		flag, err := r.IsFavorite(ctx, userID, eventID)
		if err != nil {
			return nil, err
		}
		flags[eventID] = flag
	}
	return flags, nil
}

func (r *PostgresRepository) FollowingFlags(ctx context.Context, followerUserID string, followedUserIDs []string) (map[string]socialmodel.FollowingFlag, error) {
	flags := make(map[string]socialmodel.FollowingFlag, len(followedUserIDs))
	for _, id := range followedUserIDs {
		flags[id] = socialmodel.FollowingFlag{UserID: id}
	}
	if followerUserID == "" {
		return flags, nil
	}

	for _, userID := range followedUserIDs {
		flag, err := r.IsFollowing(ctx, followerUserID, userID)
		if err != nil {
			return nil, err
		}
		flags[userID] = flag
	}
	return flags, nil
}

func scanUserFollows(rows pgx.Rows, err error) ([]socialmodel.UserFollow, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []socialmodel.UserFollow
	for rows.Next() {
		var item socialmodel.UserFollow
		if err := rows.Scan(&item.UserID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return platformerrors.ErrInvalidReference
		case "23514":
			return platformerrors.ErrInvalidReference
		}
	}
	return err
}
