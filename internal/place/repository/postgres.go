package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	placemodel "cityhawk/backend/internal/place/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

type eventProjectionRow struct {
	ID               string
	Title            string
	Categories       []string
	LikeCount        int
	ShortDescription string
	FullDescription  string
	Address          string
	ImageURL         string
	WorkingHours     string
	PriceLevel       string
}

const listEventsBaseQueryTemplate = `
	WITH first_session AS (
		SELECT DISTINCT ON (es.event_id)
			es.event_id,
			p.address_line,
			to_char(es.start_at, 'YYYY-MM-DD HH24:MI') || ' - ' || to_char(es.end_at, 'YYYY-MM-DD HH24:MI') AS working_hours,
			CASE
				WHEN es.price = 0 THEN 'free'
				WHEN es.price <= 1000 THEN 'low'
				WHEN es.price <= 3000 THEN 'medium'
				ELSE 'high'
			END AS price_level
		FROM event_session es
		JOIN place p ON p.id = es.place_id
		ORDER BY es.event_id, es.start_at ASC, es.id ASC
	),
	first_image AS (
		SELECT DISTINCT ON (ei.event_id)
			ei.event_id,
			ei.image_url
		FROM event_image ei
		ORDER BY ei.event_id, ei.created_at ASC, ei.id ASC
	),
	categories AS (
		SELECT
			ec.event_id,
			ARRAY_AGG(c.name ORDER BY c.name) AS categories
		FROM event_category ec
		JOIN category c ON c.id = ec.category_id
		GROUP BY ec.event_id
	),
	favorites AS (
		SELECT
			fe.event_id,
			COUNT(*)::int AS like_count
		FROM favorite_event fe
		GROUP BY fe.event_id
	)
	SELECT
		e.id,
		e.title,
		COALESCE(cats.categories, '{}'::text[]) AS categories,
		COALESCE(fav.like_count, 0) AS like_count,
		e.location_description,
		COALESCE(e.full_description, '') AS full_description,
		COALESCE(sess.address_line, '') AS address,
		COALESCE(img.image_url, '') AS image_url,
		COALESCE(sess.working_hours, '') AS working_hours,
		COALESCE(sess.price_level, 'unknown') AS price_level
	FROM event e
	LEFT JOIN first_session sess ON sess.event_id = e.id
	LEFT JOIN first_image img ON img.event_id = e.id
	LEFT JOIN categories cats ON cats.event_id = e.id
	LEFT JOIN favorites fav ON fav.event_id = e.id
	%s
	%s
	%s
`

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListCards(ctx context.Context) []placemodel.PlaceCard {
	rows, err := r.pool.Query(ctx, listEventsBaseQuery(listQueryOptions{}))
	if err != nil {
		return []placemodel.PlaceCard{}
	}
	defer rows.Close()

	items := make([]placemodel.PlaceCard, 0)
	for rows.Next() {
		item, err := scanEventProjectionRow(rows)
		if err != nil {
			return []placemodel.PlaceCard{}
		}
		items = append(items, toPlaceCard(item))
	}
	if rows.Err() != nil {
		return []placemodel.PlaceCard{}
	}

	return items
}

func (r *PostgresRepository) ListHomeCards(ctx context.Context) []placemodel.HomePlaceCard {
	rows, err := r.pool.Query(ctx, listEventsBaseQuery(listQueryOptions{
		OrderBy: "COALESCE(fav.like_count, 0) DESC, e.created_at DESC, e.id ASC",
		Limit:   8,
	}))
	if err != nil {
		return []placemodel.HomePlaceCard{}
	}
	defer rows.Close()

	items := make([]placemodel.HomePlaceCard, 0)
	for rows.Next() {
		item, err := scanEventProjectionRow(rows)
		if err != nil {
			return []placemodel.HomePlaceCard{}
		}
		items = append(items, placemodel.HomePlaceCard{
			ID:          item.ID,
			ImageURL:    item.ImageURL,
			Title:       item.Title,
			Description: item.ShortDescription,
		})
	}
	if rows.Err() != nil {
		return []placemodel.HomePlaceCard{}
	}

	return items
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (placemodel.Place, bool) {
	row := r.pool.QueryRow(
		ctx,
		listEventsBaseQuery(listQueryOptions{
			Where:   "e.id = $1",
			OrderBy: "e.id ASC",
			Limit:   1,
		}),
		id,
	)

	item, err := scanEventProjectionRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return placemodel.Place{}, false
		}
		return placemodel.Place{}, false
	}

	return toPlace(item), true
}

func (r *PostgresRepository) ListCardsByCategory(ctx context.Context, category string) ([]placemodel.PlaceCard, bool) {
	rows, err := r.pool.Query(
		ctx,
		listEventsBaseQuery(listQueryOptions{
			Where: `
				EXISTS (
					SELECT 1
					FROM event_category ec2
					JOIN category c2 ON c2.id = ec2.category_id
					WHERE ec2.event_id = e.id AND c2.name = $1
				)`,
			OrderBy: "e.created_at DESC, e.id ASC",
		}),
		category,
	)
	if err != nil {
		return []placemodel.PlaceCard{}, false
	}
	defer rows.Close()

	items := make([]placemodel.PlaceCard, 0)
	for rows.Next() {
		item, err := scanEventProjectionRow(rows)
		if err != nil {
			return []placemodel.PlaceCard{}, false
		}
		items = append(items, toPlaceCard(item))
	}
	if rows.Err() != nil {
		return []placemodel.PlaceCard{}, false
	}

	return items, len(items) > 0
}

type listQueryOptions struct {
	Where   string
	OrderBy string
	Limit   int
}

func listEventsBaseQuery(opts listQueryOptions) string {
	whereClause := ""
	if strings.TrimSpace(opts.Where) != "" {
		whereClause = "\nWHERE " + strings.TrimSpace(opts.Where)
	}

	orderByClause := "\nORDER BY e.created_at DESC, e.id ASC"
	if strings.TrimSpace(opts.OrderBy) != "" {
		orderByClause = "\nORDER BY " + strings.TrimSpace(opts.OrderBy)
	}

	limitClause := ""
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("\nLIMIT %d", opts.Limit)
	}

	return fmt.Sprintf(listEventsBaseQueryTemplate, whereClause, orderByClause, limitClause)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEventProjectionRow(scanner rowScanner) (eventProjectionRow, error) {
	var item eventProjectionRow
	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&item.Categories,
		&item.LikeCount,
		&item.ShortDescription,
		&item.FullDescription,
		&item.Address,
		&item.ImageURL,
		&item.WorkingHours,
		&item.PriceLevel,
	)
	return item, err
}

func toPlace(item eventProjectionRow) placemodel.Place {
	return placemodel.Place{
		ID:               item.ID,
		Title:            item.Title,
		Categories:       item.Categories,
		LikeCount:        item.LikeCount,
		ShortDescription: item.ShortDescription,
		FullDescription:  item.FullDescription,
		Address:          item.Address,
		ImageURL:         item.ImageURL,
		WorkingHours:     item.WorkingHours,
		PriceLevel:       item.PriceLevel,
	}
}

func toPlaceCard(item eventProjectionRow) placemodel.PlaceCard {
	return placemodel.PlaceCard{
		ID:               item.ID,
		Title:            item.Title,
		Categories:       item.Categories,
		LikeCount:        item.LikeCount,
		ShortDescription: item.ShortDescription,
		Address:          item.Address,
		ImageURL:         item.ImageURL,
	}
}
