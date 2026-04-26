package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	platformerrors "cityhawk/backend/internal/platform/errors"
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
		INSERT INTO user_account (email, username, user_surname, password_hash, birthday, city_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, username, user_surname, password_hash, birthday, city_id, avatar_url, role, created_at, updated_at
	`
	getUserByEmailQuery = `
		SELECT
			u.id, u.email, u.username, u.user_surname, u.password_hash, u.birthday, u.city_id, u.avatar_url, u.role, u.created_at, u.updated_at,
			c.id, c.name, c.country_name, c.timezone
		FROM user_account u
		LEFT JOIN city c ON c.id = u.city_id
		WHERE u.email = $1
	`
	getUserByIDQuery = `
		SELECT
			u.id, u.email, u.username, u.user_surname, u.password_hash, u.birthday, u.city_id, u.avatar_url, u.role, u.created_at, u.updated_at,
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
	row := r.pool.QueryRow(
		ctx,
		insertUserQuery,
		u.Email,
		u.Username,
		u.UserSurname,
		u.PasswordHash,
		u.Birthday,
		u.CityID,
		u.AvatarURL,
	)

	var persisted usermodel.User
	err := row.Scan(
		&persisted.ID,
		&persisted.Email,
		&persisted.Username,
		&persisted.UserSurname,
		&persisted.PasswordHash,
		&persisted.Birthday,
		&persisted.CityID,
		&persisted.AvatarURL,
		&persisted.Role,
		&persisted.CreatedAt,
		&persisted.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return usermodel.User{}, platformerrors.ErrEmailExists
		}
		return usermodel.User{}, err
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
	setClauses := make([]string, 0, 6)
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

	if len(setClauses) == 0 {
		u, ok := r.GetByID(ctx, id)
		return u, ok, nil
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE user_account
		SET %s, updated_at = now()
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argPos)

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return usermodel.User{}, false, platformerrors.ErrInvalidCity
			case "23505":
				return usermodel.User{}, false, platformerrors.ErrEmailExists
			}
		}
		return usermodel.User{}, false, err
	}

	u, ok := r.GetByID(ctx, id)
	return u, ok, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userScanner) (usermodel.User, error) {
	var u usermodel.User
	var cityID sql.NullString
	var avatarURL sql.NullString
	var birthday sql.NullTime
	var cityRecordID sql.NullString
	var cityName sql.NullString
	var countryName sql.NullString
	var timezone sql.NullString

	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.UserSurname,
		&u.PasswordHash,
		&birthday,
		&cityID,
		&avatarURL,
		&u.Role,
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
	if u.Role == "" {
		u.Role = usermodel.RoleUser
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
