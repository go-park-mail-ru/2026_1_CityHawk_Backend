package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	platformerrors "cityhawk/backend/internal/platform/errors"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
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
	`, userID, eventID)
	return mapPgError(err)
}

func (r *PostgresRepository) RemoveFavorite(ctx context.Context, userID, eventID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM favorite_event WHERE user_id = $1 AND event_id = $2`, userID, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	if ok, err := r.eventExists(ctx, eventID); err != nil {
		return err
	} else if !ok {
		return platformerrors.ErrInvalidReference
	}
	return nil
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

func (r *PostgresRepository) CountFavoriteEvents(ctx context.Context, userID string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM favorite_event WHERE user_id = $1`, userID).Scan(&total)
	return total, err
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
	`, followerUserID, followedUserID)
	return mapPgError(err)
}

func (r *PostgresRepository) UnfollowUser(ctx context.Context, followerUserID, followedUserID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM user_follow WHERE follower_user_id = $1 AND followed_user_id = $2`, followerUserID, followedUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	if ok, err := r.userExists(ctx, followedUserID); err != nil {
		return err
	} else if !ok {
		return platformerrors.ErrInvalidReference
	}
	return nil
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

func (r *PostgresRepository) ListFollowerProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error) {
	return r.listProfiles(ctx, viewerID, `
		FROM user_follow uf
		JOIN user_account u ON u.id = uf.follower_user_id
		LEFT JOIN city c ON c.id = u.city_id
		WHERE uf.followed_user_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, limit, offset)
}

func (r *PostgresRepository) ListFollowingProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error) {
	return r.listProfiles(ctx, viewerID, `
		FROM user_follow uf
		JOIN user_account u ON u.id = uf.followed_user_id
		LEFT JOIN city c ON c.id = u.city_id
		WHERE uf.follower_user_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, limit, offset)
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

func (r *PostgresRepository) UserCollections(ctx context.Context, userID string, limit, offset int) ([]socialmodel.CollectionCard, int, error) {
	rows, err := r.pool.Query(ctx, `
		WITH first_image AS (
			SELECT DISTINCT ON (ci.collection_id)
				ci.collection_id,
				ci.image_url
			FROM collection_image ci
			ORDER BY ci.collection_id, ci.created_at ASC, ci.id ASC
		)
		SELECT
			c.id::text,
			c.title,
			COALESCE(c.description, ''),
			COALESCE(fi.image_url, ''),
			c.is_public,
			count(*) OVER() AS total_count
		FROM collection c
		LEFT JOIN first_image fi ON fi.collection_id = c.id
		WHERE c.author_user_id = $1
		ORDER BY c.created_at DESC, c.id ASC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]socialmodel.CollectionCard, 0)
	total := 0
	for rows.Next() {
		var item socialmodel.CollectionCard
		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.ImageURL, &item.IsPublic, &total); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *PostgresRepository) listProfiles(ctx context.Context, viewerID, fromSQL string, subjectUserID string, limit, offset int) ([]socialmodel.UserProfile, int, error) {
	query := `
		SELECT
			u.id::text,
			u.username,
			u.user_surname,
			u.avatar_url,
			c.id::text,
			c.name,
			c.country_name,
			c.timezone,
			EXISTS (
				SELECT 1
				FROM user_follow vf
				WHERE vf.follower_user_id = $2
					AND vf.followed_user_id = u.id
			) AS is_following,
			count(*) OVER() AS total_count
	` + fromSQL
	rows, err := r.pool.Query(ctx, query, subjectUserID, nullableUUID(viewerID), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]socialmodel.UserProfile, 0)
	total := 0
	for rows.Next() {
		var item socialmodel.UserProfile
		var avatarURL sql.NullString
		var cityID sql.NullString
		var cityName sql.NullString
		var countryName sql.NullString
		var timezone sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Username,
			&item.UserSurname,
			&avatarURL,
			&cityID,
			&cityName,
			&countryName,
			&timezone,
			&item.IsFollowing,
			&total,
		); err != nil {
			return nil, 0, err
		}
		if avatarURL.Valid {
			value := avatarURL.String
			item.AvatarURL = &value
		}
		if cityID.Valid {
			item.City = &socialmodel.City{
				ID:          cityID.String,
				Name:        cityName.String,
				CountryName: countryName.String,
				Timezone:    timezone.String,
			}
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func nullableUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (r *PostgresRepository) eventExists(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM event WHERE id::text = $1)`, eventID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) userExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM user_account WHERE id::text = $1)`, userID).Scan(&exists)
	return exists, err
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
		case platformpostgres.CodeUniqueViolation:
			return platformerrors.ErrAlreadyExists
		case platformpostgres.CodeForeignKeyViolation:
			return platformerrors.ErrInvalidReference
		case platformpostgres.CodeCheckViolation:
			return platformerrors.ErrInvalidReference
		}
	}
	return err
}
