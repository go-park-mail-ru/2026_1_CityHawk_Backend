package repository

import (
	"context"
	"errors"

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
		INSERT INTO user_account (email, username, user_surname, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, username, password_hash
	`
	getUserByEmailQuery = `
		SELECT id, email, username, password_hash
		FROM user_account
		WHERE email = $1
	`
	getUserByIDQuery = `
		SELECT id, email, username, password_hash
		FROM user_account
		WHERE id = $1
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
		u.Username,
		u.PasswordHash,
	)

	var persisted usermodel.User
	err := row.Scan(&persisted.ID, &persisted.Email, &persisted.Username, &persisted.PasswordHash)
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

	var u usermodel.User
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash)
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

	var u usermodel.User
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodel.User{}, false
		}
		return usermodel.User{}, false
	}

	return u, true
}
