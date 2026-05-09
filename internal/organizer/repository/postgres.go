package repository

import (
	"context"
	"errors"

	organizermodel "cityhawk/backend/internal/organizer/model"
	organizerusecase "cityhawk/backend/internal/organizer/usecase"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
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

func (r *PostgresRepository) Create(ctx context.Context, input organizerusecase.CreateInput) (organizermodel.Application, error) {
	var item organizermodel.Application
	err := r.pool.QueryRow(ctx, `
		INSERT INTO organizer_application (
			user_id, status, name, email, phone, city, project_name, categories, links, about
		)
		VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text, user_id::text, status, name, email, phone, city, project_name, categories,
			COALESCE(links, ''), about, COALESCE(review_comment, ''), created_at, updated_at
	`, input.UserID, input.Name, input.Email, input.Phone, input.City, input.ProjectName, input.Categories, nullIfEmpty(input.Links), input.About).Scan(
		&item.ID,
		&item.UserID,
		&item.Status,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.City,
		&item.ProjectName,
		&item.Categories,
		&item.Links,
		&item.About,
		&item.ReviewComment,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return organizermodel.Application{}, mapPgError(err)
	}
	return item, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) (organizermodel.Application, bool, error) {
	item, err := r.queryOne(ctx, `
		SELECT id::text, user_id::text, status, name, email, phone, city, project_name, categories,
			COALESCE(links, ''), about, COALESCE(review_comment, ''), created_at, updated_at
		FROM organizer_application
		WHERE user_id::text = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return organizermodel.Application{}, false, nil
	}
	if err != nil {
		return organizermodel.Application{}, false, err
	}
	return item, true, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, input organizerusecase.StatusInput) (organizermodel.Application, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return organizermodel.Application{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	item, err := queryOne(ctx, tx, `
		UPDATE organizer_application
		SET status = $2, review_comment = $3
		WHERE id::text = $1
		RETURNING id::text, user_id::text, status, name, email, phone, city, project_name, categories,
			COALESCE(links, ''), about, COALESCE(review_comment, ''), created_at, updated_at
	`, input.ApplicationID, input.Status, nullIfEmpty(input.ReviewComment))
	if errors.Is(err, pgx.ErrNoRows) {
		return organizermodel.Application{}, false, nil
	}
	if err != nil {
		return organizermodel.Application{}, false, err
	}

	if item.Status == organizermodel.StatusApproved {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_role (user_id, role)
			VALUES ($1, 'organizer')
			ON CONFLICT DO NOTHING
		`, item.UserID); err != nil {
			return organizermodel.Application{}, false, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO organizer_profile (user_id, display_name, description, website_url, is_verified)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (user_id) DO UPDATE
			SET display_name = EXCLUDED.display_name,
				description = EXCLUDED.description,
				website_url = EXCLUDED.website_url,
				is_verified = true
		`, item.UserID, item.ProjectName, nullIfEmpty(item.About), nullIfEmpty(item.Links)); err != nil {
			return organizermodel.Application{}, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return organizermodel.Application{}, false, err
	}
	return item, true, nil
}

func (r *PostgresRepository) queryOne(ctx context.Context, sql string, args ...any) (organizermodel.Application, error) {
	return queryOne(ctx, r.pool, sql, args...)
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func queryOne(ctx context.Context, db rowQuerier, sql string, args ...any) (organizermodel.Application, error) {
	var item organizermodel.Application
	err := db.QueryRow(ctx, sql, args...).Scan(
		&item.ID,
		&item.UserID,
		&item.Status,
		&item.Name,
		&item.Email,
		&item.Phone,
		&item.City,
		&item.ProjectName,
		&item.Categories,
		&item.Links,
		&item.About,
		&item.ReviewComment,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == platformpostgres.CodeUniqueViolation {
		return organizerusecase.ErrActiveApplicationExists
	}
	return err
}
