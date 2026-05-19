package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	platformerrors "cityhawk/backend/internal/platform/errors"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	usermodel "cityhawk/backend/internal/user/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

const (
	insertUserQuery = `
		INSERT INTO user_account (email, username, user_surname, password_hash, birthday, city_id, avatar_url, bio)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	getUserByEmailQuery = `
		SELECT
			u.id, u.email, u.username, u.user_surname, u.password_hash, u.birthday, u.city_id, u.avatar_url, u.bio,
			COALESCE((
				SELECT CASE
					WHEN bool_or(ur.role = 'admin') THEN 'admin'
					WHEN bool_or(ur.role = 'organizer') THEN 'organizer'
					ELSE 'user'
				END
				FROM user_role ur
				WHERE ur.user_id = u.id
			), 'user') AS role,
			COALESCE((
				SELECT jsonb_agg(uit.tag_id::text ORDER BY uit.tag_id::text)
				FROM user_interest_tag uit
				WHERE uit.user_id = u.id
			), '[]'::jsonb)::text AS interest_tag_ids,
			u.created_at, u.updated_at,
			c.id, c.name, c.country_name, c.timezone
		FROM user_account u
		LEFT JOIN city c ON c.id = u.city_id
		WHERE u.email = $1
	`
	getUserByIDQuery = `
		SELECT
			u.id, u.email, u.username, u.user_surname, u.password_hash, u.birthday, u.city_id, u.avatar_url, u.bio,
			COALESCE((
				SELECT CASE
					WHEN bool_or(ur.role = 'admin') THEN 'admin'
					WHEN bool_or(ur.role = 'organizer') THEN 'organizer'
					ELSE 'user'
				END
				FROM user_role ur
				WHERE ur.user_id = u.id
			), 'user') AS role,
			COALESCE((
				SELECT jsonb_agg(uit.tag_id::text ORDER BY uit.tag_id::text)
				FROM user_interest_tag uit
				WHERE uit.user_id = u.id
			), '[]'::jsonb)::text AS interest_tag_ids,
			u.created_at, u.updated_at,
			c.id, c.name, c.country_name, c.timezone
		FROM user_account u
		LEFT JOIN city c ON c.id = u.city_id
		WHERE u.id = $1
	`
)

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u usermodel.User) (usermodel.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return usermodel.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id string
	err = tx.QueryRow(
		ctx,
		insertUserQuery,
		u.Email,
		u.Username,
		u.UserSurname,
		u.PasswordHash,
		u.Birthday,
		u.CityID,
		u.AvatarURL,
		u.Bio,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == platformpostgres.CodeUniqueViolation {
			return usermodel.User{}, platformerrors.ErrEmailExists
		}
		return usermodel.User{}, err
	}

	role := u.Role
	if role == "" {
		role = usermodel.RoleUser
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_role (user_id, role)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, id, role); err != nil {
		return usermodel.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return usermodel.User{}, err
	}

	persisted, ok := r.GetByID(ctx, id)
	if !ok {
		return usermodel.User{}, pgx.ErrNoRows
	}
	return persisted, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (usermodel.User, bool) {
	row := r.pool.QueryRow(
		ctx,
		getUserByEmailQuery,
		email,
	)

	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodel.User{}, false
		}
		return usermodel.User{}, false
	}

	return u, true
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (usermodel.User, bool) {
	row := r.pool.QueryRow(
		ctx,
		getUserByIDQuery,
		id,
	)

	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodel.User{}, false
		}
		return usermodel.User{}, false
	}

	return u, true
}

func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error) {
	setClauses := make([]string, 0, 7)
	args := make([]any, 0, 7)
	argPos := 1

	if patch.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *patch.Email)
		argPos++
	}
	if patch.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argPos))
		args = append(args, *patch.Username)
		argPos++
	}
	if patch.UserSurname != nil {
		setClauses = append(setClauses, fmt.Sprintf("user_surname = $%d", argPos))
		args = append(args, *patch.UserSurname)
		argPos++
	}
	if patch.Birthday != nil {
		setClauses = append(setClauses, fmt.Sprintf("birthday = $%d", argPos))
		args = append(args, patch.Birthday.UTC())
		argPos++
	}
	if patch.CityID != nil {
		setClauses = append(setClauses, fmt.Sprintf("city_id = $%d", argPos))
		args = append(args, *patch.CityID)
		argPos++
	}
	if patch.AvatarURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", argPos))
		args = append(args, *patch.AvatarURL)
		argPos++
	}
	if patch.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argPos))
		args = append(args, *patch.Bio)
		argPos++
	}

	if len(setClauses) == 0 && patch.InterestTagIDs == nil {
		u, ok := r.GetByID(ctx, id)
		return u, ok, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return usermodel.User{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if len(setClauses) > 0 {
		args = append(args, id)
		query := fmt.Sprintf(`
			UPDATE user_account
			SET %s, updated_at = now()
			WHERE id = $%d
		`, strings.Join(setClauses, ", "), argPos)

		if tag, err := tx.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case platformpostgres.CodeForeignKeyViolation:
					return usermodel.User{}, false, platformerrors.ErrInvalidCity
				case platformpostgres.CodeUniqueViolation:
					return usermodel.User{}, false, platformerrors.ErrEmailExists
				}
			}
			return usermodel.User{}, false, err
		} else if tag.RowsAffected() == 0 {
			return usermodel.User{}, false, nil
		}
	}

	if patch.InterestTagIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM user_interest_tag WHERE user_id = $1`, id); err != nil {
			return usermodel.User{}, false, err
		}
		for _, tagID := range *patch.InterestTagIDs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO user_interest_tag (user_id, tag_id)
				VALUES ($1, $2)
			`, id, tagID); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == platformpostgres.CodeForeignKeyViolation {
					return usermodel.User{}, false, platformerrors.ErrInvalidReference
				}
				return usermodel.User{}, false, err
			}
		}
		if len(setClauses) == 0 {
			if tag, err := tx.Exec(ctx, `UPDATE user_account SET updated_at = now() WHERE id = $1`, id); err != nil {
				return usermodel.User{}, false, err
			} else if tag.RowsAffected() == 0 {
				return usermodel.User{}, false, nil
			}
		}
	}

	u, err := scanUser(tx.QueryRow(ctx, getUserByIDQuery, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodel.User{}, false, nil
		}
		return usermodel.User{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return usermodel.User{}, false, err
	}
	return u, true, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userScanner) (usermodel.User, error) {
	var u usermodel.User
	var cityID sql.NullString
	var avatarURL sql.NullString
	var bio sql.NullString
	var birthday sql.NullTime
	var cityRecordID sql.NullString
	var cityName sql.NullString
	var countryName sql.NullString
	var timezone sql.NullString
	var interestTagIDs sql.NullString

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.UserSurname,
		&u.PasswordHash,
		&birthday,
		&cityID,
		&avatarURL,
		&bio,
		&u.Role,
		&interestTagIDs,
		&u.CreatedAt,
		&u.UpdatedAt,
		&cityRecordID,
		&cityName,
		&countryName,
		&timezone,
	)
	if err != nil {
		return usermodel.User{}, err
	}

	if birthday.Valid {
		t := birthday.Time.UTC()
		u.Birthday = &t
	}
	if cityID.Valid {
		v := cityID.String
		u.CityID = &v
	}
	if avatarURL.Valid {
		v := avatarURL.String
		u.AvatarURL = &v
	}
	if bio.Valid {
		v := bio.String
		u.Bio = &v
	}
	if u.Role == "" {
		u.Role = usermodel.RoleUser
	}
	if interestTagIDs.Valid {
		tagIDs, err := parseInterestTagIDs(interestTagIDs.String)
		if err != nil {
			return usermodel.User{}, err
		}
		u.InterestTagIDs = tagIDs
	}
	if cityRecordID.Valid {
		u.City = &usermodel.City{
			ID:          cityRecordID.String,
			Name:        cityName.String,
			CountryName: countryName.String,
			Timezone:    timezone.String,
		}
	}

	return u, nil
}

func parseInterestTagIDs(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}

	var tagIDs []string
	if err := json.Unmarshal([]byte(raw), &tagIDs); err != nil {
		return nil, fmt.Errorf("parse interest tag ids: %w", err)
	}
	out := make([]string, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		value := strings.TrimSpace(tagID)
		if value != "" {
			out = append(out, value)
		}
	}
	return out, nil
}
