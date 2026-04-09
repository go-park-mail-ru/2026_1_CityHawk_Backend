package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

type homeFeaturedEventRow struct {
	ID            string
	Title         string
	CoverImageURL string
	TagIDs        []string
	TagNames      []string
	StartAt       time.Time
	PlaceName     string
	AddressLine   string
}

type homeCategoryRow struct {
	ID   string
	Name string
}

type homeCollectionRow struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
}

type eventProjectionRow struct {
	ID                  string
	Title               string
	Categories          []string
	LikeCount           int
	LocationDescription string
	FullDescription     string
	Address             string
	ImageURL            string
	WorkingHours        string
	PriceLevel          string
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

func (r *PostgresRepository) ListFeaturedEvents(ctx context.Context, limit int) []placemodel.HomeFeaturedEvent {
	if limit <= 0 {
		limit = 8
	}

	const query = `
		WITH first_image AS (
			SELECT DISTINCT ON (ei.event_id)
				ei.event_id,
				ei.image_url
			FROM event_image ei
			ORDER BY ei.event_id, ei.created_at ASC, ei.id ASC
		),
		next_session AS (
			SELECT DISTINCT ON (es.event_id)
				es.event_id,
				es.start_at,
				p.name AS place_name,
				p.address_line
			FROM event_session es
			JOIN place p ON p.id = es.place_id
			ORDER BY es.event_id, es.start_at ASC, es.id ASC
		),
		tags AS (
			SELECT
				et.event_id,
				ARRAY_AGG(t.id::text ORDER BY t.name) AS tag_ids,
				ARRAY_AGG(t.name ORDER BY t.name) AS tag_names
			FROM event_tag et
			JOIN tag t ON t.id = et.tag_id
			GROUP BY et.event_id
		)
		SELECT
			e.id::text,
			e.title,
			COALESCE(fi.image_url, '') AS cover_image_url,
			COALESCE(tags.tag_ids, '{}'::text[]) AS tag_ids,
			COALESCE(tags.tag_names, '{}'::text[]) AS tag_names,
			ns.start_at,
			COALESCE(ns.place_name, '') AS place_name,
			COALESCE(ns.address_line, '') AS address_line
		FROM event e
		LEFT JOIN first_image fi ON fi.event_id = e.id
		LEFT JOIN next_session ns ON ns.event_id = e.id
		LEFT JOIN tags ON tags.event_id = e.id
		ORDER BY e.created_at DESC, e.id ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return []placemodel.HomeFeaturedEvent{}
	}
	defer rows.Close()

	items := make([]placemodel.HomeFeaturedEvent, 0, limit)
	for rows.Next() {
		var row homeFeaturedEventRow
		if err := rows.Scan(
			&row.ID,
			&row.Title,
			&row.CoverImageURL,
			&row.TagIDs,
			&row.TagNames,
			&row.StartAt,
			&row.PlaceName,
			&row.AddressLine,
		); err != nil {
			return []placemodel.HomeFeaturedEvent{}
		}
		items = append(items, toHomeFeaturedEvent(row))
	}
	if rows.Err() != nil {
		return []placemodel.HomeFeaturedEvent{}
	}

	return items
}

func (r *PostgresRepository) ListHomeCategories(ctx context.Context, limit int) []placemodel.HomeCategory {
	if limit <= 0 {
		limit = 8
	}

	const query = `
		SELECT c.id::text, c.name
		FROM category c
		ORDER BY c.name ASC, c.id ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return []placemodel.HomeCategory{}
	}
	defer rows.Close()

	items := make([]placemodel.HomeCategory, 0, limit)
	for rows.Next() {
		var row homeCategoryRow
		if err := rows.Scan(&row.ID, &row.Name); err != nil {
			return []placemodel.HomeCategory{}
		}
		items = append(items, placemodel.HomeCategory{
			ID:   row.ID,
			Name: row.Name,
			Slug: slugify(row.Name),
		})
	}
	if rows.Err() != nil {
		return []placemodel.HomeCategory{}
	}

	return items
}

func (r *PostgresRepository) ListHomeCollections(ctx context.Context, limit int) []placemodel.HomeCollection {
	if limit <= 0 {
		limit = 8
	}

	const query = `
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
			COALESCE(c.description, '') AS description,
			COALESCE(fi.image_url, '') AS image_url
		FROM collection c
		LEFT JOIN first_image fi ON fi.collection_id = c.id
		ORDER BY c.created_at DESC, c.id ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return []placemodel.HomeCollection{}
	}
	defer rows.Close()

	items := make([]placemodel.HomeCollection, 0, limit)
	for rows.Next() {
		var row homeCollectionRow
		if err := rows.Scan(&row.ID, &row.Title, &row.Description, &row.ImageURL); err != nil {
			return []placemodel.HomeCollection{}
		}
		items = append(items, placemodel.HomeCollection(row))
	}
	if rows.Err() != nil {
		return []placemodel.HomeCollection{}
	}

	return items
}

func (r *PostgresRepository) ListCards(ctx context.Context) []placemodel.EventCardView {
	rows, err := r.pool.Query(ctx, listEventsBaseQuery(listQueryOptions{}))
	if err != nil {
		return []placemodel.EventCardView{}
	}
	defer rows.Close()

	items := make([]placemodel.EventCardView, 0)
	for rows.Next() {
		item, err := scanEventProjectionRow(rows)
		if err != nil {
			return []placemodel.EventCardView{}
		}
		items = append(items, toPlaceCard(item))
	}
	if rows.Err() != nil {
		return []placemodel.EventCardView{}
	}

	return items
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (placemodel.EventDetailsView, bool) {
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
			return placemodel.EventDetailsView{}, false
		}
		return placemodel.EventDetailsView{}, false
	}

	return toPlace(item), true
}

func (r *PostgresRepository) ListCardsByCategory(ctx context.Context, category string) ([]placemodel.EventCardView, bool) {
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
		return []placemodel.EventCardView{}, false
	}
	defer rows.Close()

	items := make([]placemodel.EventCardView, 0)
	for rows.Next() {
		item, err := scanEventProjectionRow(rows)
		if err != nil {
			return []placemodel.EventCardView{}, false
		}
		items = append(items, toPlaceCard(item))
	}
	if rows.Err() != nil {
		return []placemodel.EventCardView{}, false
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
		&item.LocationDescription,
		&item.FullDescription,
		&item.Address,
		&item.ImageURL,
		&item.WorkingHours,
		&item.PriceLevel,
	)
	return item, err
}

func toPlace(item eventProjectionRow) placemodel.EventDetailsView {
	return placemodel.EventDetailsView{
		ID:                  item.ID,
		Title:               item.Title,
		Categories:          item.Categories,
		LikeCount:           item.LikeCount,
		LocationDescription: item.LocationDescription,
		FullDescription:     item.FullDescription,
		AddressLine:         item.Address,
		ImageURL:            item.ImageURL,
		SessionLabel:        item.WorkingHours,
		PriceLevel:          item.PriceLevel,
	}
}

func toPlaceCard(item eventProjectionRow) placemodel.EventCardView {
	return placemodel.EventCardView{
		ID:                  item.ID,
		Title:               item.Title,
		Categories:          item.Categories,
		LikeCount:           item.LikeCount,
		LocationDescription: item.LocationDescription,
		AddressLine:         item.Address,
		ImageURL:            item.ImageURL,
	}
}

func toHomeFeaturedEvent(row homeFeaturedEventRow) placemodel.HomeFeaturedEvent {
	tags := make([]placemodel.HomeTag, 0, len(row.TagNames))
	for i, name := range row.TagNames {
		id := ""
		if i < len(row.TagIDs) {
			id = row.TagIDs[i]
		}
		tags = append(tags, placemodel.HomeTag{
			ID:   id,
			Name: name,
			Slug: slugify(name),
		})
	}

	return placemodel.HomeFeaturedEvent{
		ID:            row.ID,
		Title:         row.Title,
		CoverImageURL: row.CoverImageURL,
		Tags:          tags,
		NextSession: placemodel.HomeNextSession{
			StartAt: row.StartAt,
			Place: placemodel.HomeNextSessionPlace{
				Name:        row.PlaceName,
				AddressLine: row.AddressLine,
			},
		},
	}
}

var slugifyRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = slugifyRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "item"
	}
	return s
}

func uniqueSortedCategories(items []placemodel.EventDetailsView, limit int) []placemodel.HomeCategory {
	set := make(map[string]struct{})
	for _, item := range items {
		for _, category := range item.Categories {
			set[category] = struct{}{}
		}
	}

	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)

	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}

	out := make([]placemodel.HomeCategory, 0, len(names))
	for _, name := range names {
		out = append(out, placemodel.HomeCategory{
			ID:   slugify(name),
			Name: name,
			Slug: slugify(name),
		})
	}
	return out
}
