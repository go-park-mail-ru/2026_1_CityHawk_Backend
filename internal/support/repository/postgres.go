package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	supportmodel "cityhawk/backend/internal/support/model"
	supportusecase "cityhawk/backend/internal/support/usecase"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, input supportusecase.CreateInput) (supportmodel.Ticket, error) {
	const query = `
		INSERT INTO support_ticket (user_id, category, status, title, message)
		VALUES ($1::uuid, $2, 'open', $3, $4)
		RETURNING id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
	`
	return r.scanTicket(r.pool.QueryRow(ctx, query, input.UserID, input.Category, input.Title, input.Message))
}

func (r *PostgresRepository) ListByUser(ctx context.Context, userID string, filter supportmodel.Filter) ([]supportmodel.Ticket, error) {
	const query = `
		SELECT id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
		FROM support_ticket
		WHERE user_id = $1::uuid
			AND ($2 = '' OR status = $2)
			AND ($3 = '' OR category = $3)
		ORDER BY created_at DESC, id DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, userID, string(filter.Status), string(filter.Category), filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]supportmodel.Ticket, 0)
	for rows.Next() {
		item, err := scanTicketRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *PostgresRepository) ListAll(ctx context.Context, filter supportmodel.Filter) ([]supportmodel.Ticket, error) {
	const query = `
		SELECT id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
		FROM support_ticket
		WHERE ($1 = '' OR status = $1)
			AND ($2 = '' OR category = $2)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, string(filter.Status), string(filter.Category), filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]supportmodel.Ticket, 0)
	for rows.Next() {
		item, err := scanTicketRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (supportmodel.Ticket, error) {
	const query = `
		SELECT id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
		FROM support_ticket
		WHERE id = $1::uuid
	`
	return r.scanTicket(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresRepository) UpdateByUser(ctx context.Context, userID string, ticketID string, input supportusecase.UpdateInput) (supportmodel.Ticket, error) {
	var (
		category any
		title    any
		message  any
	)
	if input.Category != nil {
		category = string(*input.Category)
	}
	if input.Title != nil {
		title = *input.Title
	}
	if input.Message != nil {
		message = *input.Message
	}

	const query = `
		UPDATE support_ticket
		SET
			category = COALESCE($3, category),
			title = COALESCE($4, title),
			message = COALESCE($5, message)
		WHERE id = $1::uuid
			AND user_id = $2::uuid
			AND status <> 'closed'
		RETURNING id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
	`
	return r.scanTicket(r.pool.QueryRow(ctx, query, ticketID, userID, category, title, message))
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, ticketID string, input supportusecase.StatusInput) (supportmodel.Ticket, error) {
	const query = `
		UPDATE support_ticket
		SET
			status = $2,
			closed_at = CASE WHEN $2 = 'closed' THEN now() ELSE NULL END
		WHERE id = $1::uuid
		RETURNING id::text, user_id::text, category, status, title, message, created_at, updated_at, closed_at
	`
	return r.scanTicket(r.pool.QueryRow(ctx, query, ticketID, input.Status))
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, input supportusecase.CreateMessageInput) (supportmodel.Message, error) {
	const query = `
		INSERT INTO support_ticket_message (ticket_id, author_user_id, author_role, body)
		VALUES ($1::uuid, $2::uuid, $3, $4)
		RETURNING id::text, ticket_id::text, author_user_id::text, author_role, body, created_at
	`
	return scanMessageRow(r.pool.QueryRow(ctx, query, input.TicketID, input.AuthorUserID, input.AuthorRole, input.Body))
}

func (r *PostgresRepository) ListMessages(ctx context.Context, ticketID string) ([]supportmodel.Message, error) {
	const query = `
		SELECT id::text, ticket_id::text, author_user_id::text, author_role, body, created_at
		FROM support_ticket_message
		WHERE ticket_id = $1::uuid
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]supportmodel.Message, 0)
	for rows.Next() {
		item, err := scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *PostgresRepository) Stats(ctx context.Context, filter supportmodel.StatsFilter) (supportmodel.Stats, error) {
	stats := supportmodel.Stats{
		ByStatus:   map[supportmodel.Status]int{},
		ByCategory: map[supportmodel.Category]int{},
	}

	statusRows, err := r.pool.Query(ctx, `
		SELECT status, count(*)::int
		FROM support_ticket
		WHERE ($1::timestamptz IS NULL OR created_at >= $1)
			AND ($2::timestamptz IS NULL OR created_at < $2)
		GROUP BY status
	`, filter.From, filter.To)
	if err != nil {
		return supportmodel.Stats{}, err
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status supportmodel.Status
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return supportmodel.Stats{}, err
		}
		stats.ByStatus[status] = count
		stats.Total += count
		switch status {
		case supportmodel.StatusOpen:
			stats.OpenTotal = count
		case supportmodel.StatusInProgress:
			stats.InProgressTotal = count
		case supportmodel.StatusClosed:
			stats.ClosedTotal = count
		}
	}
	if statusRows.Err() != nil {
		return supportmodel.Stats{}, statusRows.Err()
	}

	categoryRows, err := r.pool.Query(ctx, `
		SELECT category, count(*)::int
		FROM support_ticket
		WHERE ($1::timestamptz IS NULL OR created_at >= $1)
			AND ($2::timestamptz IS NULL OR created_at < $2)
		GROUP BY category
	`, filter.From, filter.To)
	if err != nil {
		return supportmodel.Stats{}, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var category supportmodel.Category
		var count int
		if err := categoryRows.Scan(&category, &count); err != nil {
			return supportmodel.Stats{}, err
		}
		stats.ByCategory[category] = count
	}
	if categoryRows.Err() != nil {
		return supportmodel.Stats{}, categoryRows.Err()
	}

	return stats, nil
}

func (r *PostgresRepository) scanTicket(row pgx.Row) (supportmodel.Ticket, error) {
	item, err := scanTicketRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return supportmodel.Ticket{}, supportusecase.ErrTicketNotFound
	}
	return item, err
}

type ticketScanner interface {
	Scan(dest ...any) error
}

func scanTicketRows(row ticketScanner) (supportmodel.Ticket, error) {
	return scanTicketRow(row)
}

func scanTicketRow(row ticketScanner) (supportmodel.Ticket, error) {
	var item supportmodel.Ticket
	var category string
	var status string
	var closedAt *time.Time
	if err := row.Scan(
		&item.ID,
		&item.UserID,
		&category,
		&status,
		&item.Title,
		&item.Message,
		&item.CreatedAt,
		&item.UpdatedAt,
		&closedAt,
	); err != nil {
		return supportmodel.Ticket{}, err
	}
	item.Category = supportmodel.Category(category)
	item.Status = supportmodel.Status(status)
	item.ClosedAt = closedAt
	return item, nil
}

func scanMessageRow(row ticketScanner) (supportmodel.Message, error) {
	var item supportmodel.Message
	var authorRole string
	if err := row.Scan(
		&item.ID,
		&item.TicketID,
		&item.AuthorUserID,
		&authorRole,
		&item.Body,
		&item.CreatedAt,
	); err != nil {
		return supportmodel.Message{}, err
	}
	item.AuthorRole = supportmodel.MessageAuthorRole(authorRole)
	return item, nil
}
