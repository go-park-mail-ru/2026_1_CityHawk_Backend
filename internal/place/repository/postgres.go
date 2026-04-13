package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

type collectionCardRow struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	IsPublic    bool
}

type searchSuggestionRow struct {
	Name string
}

type eventListRow struct {
	ID            string
	Title         string
	ShortDesc     string
	CoverImageURL string
	TagIDs        []string
	TagNames      []string
	StartAt       *time.Time
	PlaceName     string
	AddressLine   string
	TotalCount    int
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) HomePayload(ctx context.Context) placemodel.HomePayload {
	var (
		featuredEvents []placemodel.HomeFeaturedEvent
		categories     []placemodel.HomeCategory
		collections    []placemodel.HomeCollection
		wg             sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		featuredEvents = r.ListFeaturedEvents(ctx, 8)
	}()

	go func() {
		defer wg.Done()
		categories = r.ListHomeCategories(ctx, 8)
	}()

	go func() {
		defer wg.Done()
		collections = r.ListHomeCollections(ctx, 8)
	}()

	wg.Wait()

	return placemodel.HomePayload{
		FeaturedEvents: featuredEvents,
		Categories:     categories,
		Collections:    collections,
	}
}

func (r *PostgresRepository) ListCategories(ctx context.Context) []placemodel.HomeCategory {
	rows, err := r.pool.Query(ctx, `SELECT c.id::text, c.name FROM category c ORDER BY c.name ASC, c.id ASC`)
	if err != nil {
		return []placemodel.HomeCategory{}
	}
	defer rows.Close()

	items := make([]placemodel.HomeCategory, 0)
	for rows.Next() {
		var row homeCategoryRow
		if err := rows.Scan(&row.ID, &row.Name); err != nil {
			return []placemodel.HomeCategory{}
		}
		items = append(items, placemodel.HomeCategory{ID: row.ID, Name: row.Name, Slug: slugify(row.Name)})
	}
	if rows.Err() != nil {
		return []placemodel.HomeCategory{}
	}
	return items
}

func (r *PostgresRepository) ListTags(ctx context.Context) []placemodel.HomeTag {
	rows, err := r.pool.Query(ctx, `SELECT t.id::text, t.name FROM tag t ORDER BY t.name ASC, t.id ASC`)
	if err != nil {
		return []placemodel.HomeTag{}
	}
	defer rows.Close()

	items := make([]placemodel.HomeTag, 0)
	for rows.Next() {
		var id string
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return []placemodel.HomeTag{}
		}
		items = append(items, placemodel.HomeTag{
			ID:   id,
			Name: name,
			Slug: slugify(name),
		})
	}
	if rows.Err() != nil {
		return []placemodel.HomeTag{}
	}
	return items
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
		if err := rows.Scan(&row.ID, &row.Title, &row.CoverImageURL, &row.TagIDs, &row.TagNames, &row.StartAt, &row.PlaceName, &row.AddressLine); err != nil {
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

	rows, err := r.pool.Query(ctx, `SELECT c.id::text, c.name FROM category c ORDER BY c.name ASC, c.id ASC LIMIT $1`, limit)
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
		items = append(items, placemodel.HomeCategory{ID: row.ID, Name: row.Name, Slug: slugify(row.Name)})
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

func (r *PostgresRepository) ListCollections(ctx context.Context) ([]placemodel.CollectionCardView, error) {
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
			COALESCE(fi.image_url, '') AS image_url,
			c.is_public
		FROM collection c
		LEFT JOIN first_image fi ON fi.collection_id = c.id
		WHERE c.is_public = TRUE
		ORDER BY c.created_at DESC, c.id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.CollectionCardView, 0)
	for rows.Next() {
		var row collectionCardRow
		if err := rows.Scan(&row.ID, &row.Title, &row.Description, &row.ImageURL, &row.IsPublic); err != nil {
			return nil, err
		}
		items = append(items, placemodel.CollectionCardView(row))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *PostgresRepository) GetCollectionByID(ctx context.Context, id string) (placemodel.CollectionDetailsView, bool, error) {
	const baseQuery = `
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
			COALESCE(fi.image_url, '') AS image_url,
			c.is_public
		FROM collection c
		LEFT JOIN first_image fi ON fi.collection_id = c.id
		WHERE c.id = $1 AND c.is_public = TRUE
	`

	var item placemodel.CollectionDetailsView
	if err := r.pool.QueryRow(ctx, baseQuery, id).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.ImageURL,
		&item.IsPublic,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return placemodel.CollectionDetailsView{}, false, nil
		}
		return placemodel.CollectionDetailsView{}, false, err
	}

	events, err := r.fetchCollectionEvents(ctx, id)
	if err != nil {
		return placemodel.CollectionDetailsView{}, false, err
	}
	item.Events = events

	return item, true, nil
}

func (r *PostgresRepository) SearchSuggestions(ctx context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error) {
	if limit <= 0 {
		limit = 5
	}

	const sqlQuery = `
		WITH candidates AS (
			SELECT
				e.title AS name,
				similarity(lower(e.title), lower($1)) AS rank
			FROM event e
			WHERE lower(e.title) LIKE lower($1) || '%'
			   OR e.title % $1
			UNION ALL
			SELECT
				c.name AS name,
				similarity(lower(c.name), lower($1)) AS rank
			FROM category c
			WHERE lower(c.name) LIKE lower($1) || '%'
			   OR c.name % $1
			UNION ALL
			SELECT
				t.name AS name,
				similarity(lower(t.name), lower($1)) AS rank
			FROM tag t
			WHERE lower(t.name) LIKE lower($1) || '%'
			   OR t.name % $1
			UNION ALL
			SELECT
				col.title AS name,
				similarity(lower(col.title), lower($1)) AS rank
			FROM collection col
			WHERE col.is_public = TRUE
			  AND (
					lower(col.title) LIKE lower($1) || '%'
					OR col.title % $1
			  )
		),
		deduped AS (
			SELECT
				name,
				MAX(rank) AS rank
			FROM candidates
			GROUP BY name
		)
		SELECT name
		FROM deduped
		ORDER BY
			CASE WHEN lower(name) LIKE lower($1) || '%' THEN 0 ELSE 1 END,
			rank DESC,
			char_length(name) ASC,
			name ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, sqlQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.SearchSuggestion, 0, limit)
	for rows.Next() {
		var row searchSuggestionRow
		if err := rows.Scan(&row.Name); err != nil {
			return nil, err
		}
		items = append(items, placemodel.SearchSuggestion{Name: row.Name})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func (r *PostgresRepository) ListEvents(ctx context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error) {
	query, args := buildListEventsQuery(filter)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]placemodel.EventCardView, 0)
	total := 0
	for rows.Next() {
		var row eventListRow
		if err := rows.Scan(&row.ID, &row.Title, &row.ShortDesc, &row.CoverImageURL, &row.TagIDs, &row.TagNames, &row.StartAt, &row.PlaceName, &row.AddressLine, &row.TotalCount); err != nil {
			return nil, 0, err
		}
		total = row.TotalCount
		items = append(items, toEventCard(row))
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return items, total, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error) {
	const baseQuery = `
		SELECT
			e.id::text,
			e.title,
			e.location_description,
			COALESCE(e.full_description, '') AS full_description,
			e.age_limit,
			e.source_url,
			u.id::text,
			u.username,
			u.avatar_url,
			e.created_at,
			e.updated_at,
			EXISTS (
				SELECT 1
				FROM favorite_event fe
				WHERE fe.event_id = e.id AND fe.user_id::text = $2
			),
			e.author_user_id::text = $2
		FROM event e
		JOIN user_account u ON u.id = e.author_user_id
		WHERE e.id = $1
	`

	var item placemodel.EventDetailsView
	var sourceURL *string
	if err := r.pool.QueryRow(ctx, baseQuery, id, userID).Scan(
		&item.ID,
		&item.Title,
		&item.ShortDescription,
		&item.FullDescription,
		&item.AgeLimit,
		&sourceURL,
		&item.Author.ID,
		&item.Author.Username,
		&item.Author.AvatarURL,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.IsFavorite,
		&item.IsOwner,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return placemodel.EventDetailsView{}, false, nil
		}
		return placemodel.EventDetailsView{}, false, err
	}
	item.SourceURL = sourceURL

	categories, err := fetchTaxonomy(ctx, r.pool, `
		SELECT c.id::text, c.name
		FROM event_category ec
		JOIN category c ON c.id = ec.category_id
		WHERE ec.event_id = $1
		ORDER BY c.name ASC, c.id ASC
	`, id)
	if err != nil {
		return placemodel.EventDetailsView{}, false, err
	}
	item.Categories = categories

	tags, err := fetchTaxonomy(ctx, r.pool, `
		SELECT t.id::text, t.name
		FROM event_tag et
		JOIN tag t ON t.id = et.tag_id
		WHERE et.event_id = $1
		ORDER BY t.name ASC, t.id ASC
	`, id)
	if err != nil {
		return placemodel.EventDetailsView{}, false, err
	}
	item.Tags = tags

	images, err := fetchImages(ctx, r.pool, id)
	if err != nil {
		return placemodel.EventDetailsView{}, false, err
	}
	item.Images = images

	sessions, err := fetchSessions(ctx, r.pool, id)
	if err != nil {
		return placemodel.EventDetailsView{}, false, err
	}
	item.Sessions = sessions

	return item, true, nil
}

func (r *PostgresRepository) CreateEvent(ctx context.Context, input placemodel.EventWriteInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer rollbackTx(ctx, tx)

	var eventID string
	var sourceURL any
	if input.SourceURL != nil {
		sourceURL = *input.SourceURL
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO event (author_user_id, title, location_description, full_description, age_limit, source_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, input.AuthorUserID, derefString(input.Title), derefString(input.ShortDescription), derefString(input.FullDescription), derefInt(input.AgeLimit), sourceURL).Scan(&eventID)
	if err != nil {
		return "", mapPGError(err)
	}

	if err := replaceCategories(ctx, tx, eventID, input.CategoryIDs); err != nil {
		return "", err
	}
	if err := replaceTags(ctx, tx, eventID, input.TagIDs); err != nil {
		return "", err
	}
	if err := replaceImages(ctx, tx, eventID, input.ImageURLs); err != nil {
		return "", err
	}
	if err := replaceSessions(ctx, tx, eventID, input.Sessions); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return eventID, nil
}

func (r *PostgresRepository) UpdateEvent(ctx context.Context, input placemodel.EventWriteInput) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer rollbackTx(ctx, tx)

	ownerID, err := getEventOwner(ctx, tx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if ownerID != input.AuthorUserID {
		return false, platformerrors.ErrForbidden
	}

	setParts := make([]string, 0, 6)
	args := make([]any, 0, 7)
	argPos := 1

	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argPos))
		args = append(args, *input.Title)
		argPos++
	}
	if input.ShortDescription != nil {
		setParts = append(setParts, fmt.Sprintf("location_description = $%d", argPos))
		args = append(args, *input.ShortDescription)
		argPos++
	}
	if input.FullDescription != nil {
		setParts = append(setParts, fmt.Sprintf("full_description = $%d", argPos))
		args = append(args, *input.FullDescription)
		argPos++
	}
	if input.AgeLimit != nil {
		setParts = append(setParts, fmt.Sprintf("age_limit = $%d", argPos))
		args = append(args, *input.AgeLimit)
		argPos++
	}
	if input.SourceURL != nil {
		setParts = append(setParts, fmt.Sprintf("source_url = $%d", argPos))
		args = append(args, *input.SourceURL)
		argPos++
	}
	if input.ClearSourceURL {
		setParts = append(setParts, "source_url = NULL")
	}

	if len(setParts) > 0 {
		args = append(args, input.ID)
		query := fmt.Sprintf(`UPDATE event SET %s, updated_at = now() WHERE id = $%d`, strings.Join(setParts, ", "), argPos)
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return false, mapPGError(err)
		}
	}

	if input.CategoryIDs != nil {
		if err := replaceCategories(ctx, tx, input.ID, input.CategoryIDs); err != nil {
			return false, err
		}
	}
	if input.TagIDs != nil {
		if err := replaceTags(ctx, tx, input.ID, input.TagIDs); err != nil {
			return false, err
		}
	}
	if input.ImageURLs != nil {
		if err := replaceImages(ctx, tx, input.ID, input.ImageURLs); err != nil {
			return false, err
		}
	}
	if input.Sessions != nil {
		if err := replaceSessions(ctx, tx, input.ID, input.Sessions); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, id, userID string) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer rollbackTx(ctx, tx)

	ownerID, err := getEventOwner(ctx, tx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if ownerID != userID {
		return false, platformerrors.ErrForbidden
	}

	tag, err := tx.Exec(ctx, `DELETE FROM event WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func buildListEventsQuery(filter placemodel.EventListFilter) (string, []any) {
	args := make([]any, 0, 10)
	whereParts := make([]string, 0, 8)
	argPos := 1

	if filter.Query != "" {
		whereParts = append(whereParts, fmt.Sprintf("(e.title ILIKE $%d OR e.location_description ILIKE $%d OR COALESCE(e.full_description, '') ILIKE $%d)", argPos, argPos, argPos))
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}
	if filter.CategoryID != "" {
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM event_category ec WHERE ec.event_id = e.id AND ec.category_id::text = $%d)", argPos))
		args = append(args, filter.CategoryID)
		argPos++
	}
	if filter.TagID != "" {
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM event_tag et WHERE et.event_id = e.id AND et.tag_id::text = $%d)", argPos))
		args = append(args, filter.TagID)
		argPos++
	}
	if filter.CityID != "" {
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM event_session es JOIN place p ON p.id = es.place_id WHERE es.event_id = e.id AND p.city_id::text = $%d)", argPos))
		args = append(args, filter.CityID)
		argPos++
	}
	if filter.AuthorID != "" {
		whereParts = append(whereParts, fmt.Sprintf("e.author_user_id::text = $%d", argPos))
		args = append(args, filter.AuthorID)
		argPos++
	}
	if filter.DateFrom != nil {
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM event_session es WHERE es.event_id = e.id AND es.start_at >= $%d)", argPos))
		args = append(args, *filter.DateFrom)
		argPos++
	}
	if filter.DateTo != nil {
		whereParts = append(whereParts, fmt.Sprintf("EXISTS (SELECT 1 FROM event_session es WHERE es.event_id = e.id AND es.start_at <= $%d)", argPos))
		args = append(args, *filter.DateTo)
		argPos++
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	orderBy := "e.created_at DESC, e.id ASC"
	switch filter.Sort {
	case "dateAsc":
		orderBy = "ns.start_at ASC NULLS LAST, e.id ASC"
	case "dateDesc":
		orderBy = "ns.start_at DESC NULLS LAST, e.id ASC"
	case "titleAsc":
		orderBy = "e.title ASC, e.id ASC"
	}

	query := fmt.Sprintf(`
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
		),
		base AS (
			SELECT
				e.id,
				e.title,
				e.location_description,
				COALESCE(fi.image_url, '') AS cover_image_url,
				COALESCE(tags.tag_ids, '{}'::text[]) AS tag_ids,
				COALESCE(tags.tag_names, '{}'::text[]) AS tag_names,
				ns.start_at,
				COALESCE(ns.place_name, '') AS place_name,
				COALESCE(ns.address_line, '') AS address_line,
				COUNT(*) OVER()::int AS total_count
			FROM event e
			LEFT JOIN first_image fi ON fi.event_id = e.id
			LEFT JOIN next_session ns ON ns.event_id = e.id
			LEFT JOIN tags ON tags.event_id = e.id
			%s
			ORDER BY %s
			LIMIT $%d OFFSET $%d
		)
		SELECT
			id::text,
			title,
			location_description,
			cover_image_url,
			tag_ids,
			tag_names,
			start_at,
			place_name,
			address_line,
			total_count
		FROM base
	`, whereClause, orderBy, argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)
	return query, args
}

func fetchTaxonomy(ctx context.Context, db queryable, query, eventID string) ([]placemodel.EventTaxonomyItem, error) {
	rows, err := db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.EventTaxonomyItem, 0)
	for rows.Next() {
		var id string
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		items = append(items, placemodel.EventTaxonomyItem{ID: id, Name: name, Slug: slugify(name)})
	}
	return items, rows.Err()
}

func fetchImages(ctx context.Context, db queryable, eventID string) ([]placemodel.EventImageView, error) {
	rows, err := db.Query(ctx, `SELECT id::text, image_url FROM event_image WHERE event_id = $1 ORDER BY created_at ASC, id ASC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.EventImageView, 0)
	for rows.Next() {
		var item placemodel.EventImageView
		if err := rows.Scan(&item.ID, &item.ImageURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func fetchSessions(ctx context.Context, db queryable, eventID string) ([]placemodel.EventSessionView, error) {
	const query = `
		SELECT
			es.id::text,
			es.start_at,
			es.end_at,
			es.price,
			p.id::text,
			p.name,
			p.address_line,
			p.latitude::float8,
			p.longitude::float8,
			c.id::text,
			c.name,
			c.country_name,
			c.timezone
		FROM event_session es
		JOIN place p ON p.id = es.place_id
		JOIN city c ON c.id = p.city_id
		WHERE es.event_id = $1
		ORDER BY es.start_at ASC, es.id ASC
	`

	rows, err := db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.EventSessionView, 0)
	for rows.Next() {
		var item placemodel.EventSessionView
		if err := rows.Scan(
			&item.ID,
			&item.StartAt,
			&item.EndAt,
			&item.Price,
			&item.Place.ID,
			&item.Place.Name,
			&item.Place.AddressLine,
			&item.Place.Latitude,
			&item.Place.Longitude,
			&item.Place.City.ID,
			&item.Place.City.Name,
			&item.Place.City.CountryName,
			&item.Place.City.Timezone,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func replaceCategories(ctx context.Context, tx pgx.Tx, eventID string, ids *[]string) error {
	if ids == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_category WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, id := range *ids {
		if _, err := tx.Exec(ctx, `INSERT INTO event_category (event_id, category_id) VALUES ($1, $2)`, eventID, id); err != nil {
			return mapPGError(err)
		}
	}
	return nil
}

func replaceTags(ctx context.Context, tx pgx.Tx, eventID string, ids *[]string) error {
	if ids == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_tag WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, id := range *ids {
		if _, err := tx.Exec(ctx, `INSERT INTO event_tag (event_id, tag_id) VALUES ($1, $2)`, eventID, id); err != nil {
			return mapPGError(err)
		}
	}
	return nil
}

func replaceImages(ctx context.Context, tx pgx.Tx, eventID string, urls *[]string) error {
	if urls == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_image WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, imageURL := range *urls {
		if _, err := tx.Exec(ctx, `INSERT INTO event_image (event_id, image_url) VALUES ($1, $2)`, eventID, imageURL); err != nil {
			return mapPGError(err)
		}
	}
	return nil
}

func replaceSessions(ctx context.Context, tx pgx.Tx, eventID string, sessions *[]placemodel.EventSessionInput) error {
	if sessions == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_session WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, session := range *sessions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
			VALUES ($1, $2, $3, $4, $5)
		`, eventID, session.PlaceID, session.StartAt, session.EndAt, session.Price); err != nil {
			return mapPGError(err)
		}
	}
	return nil
}

func getEventOwner(ctx context.Context, tx pgx.Tx, eventID string) (string, error) {
	var ownerID string
	err := tx.QueryRow(ctx, `SELECT author_user_id::text FROM event WHERE id = $1`, eventID).Scan(&ownerID)
	return ownerID, err
}

func mapPGError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return platformerrors.ErrInvalidReference
		}
	}
	return err
}

type queryable interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func rollbackTx(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func toEventCard(row eventListRow) placemodel.EventCardView {
	tags := make([]placemodel.EventTaxonomyItem, 0, len(row.TagNames))
	for i, name := range row.TagNames {
		id := ""
		if i < len(row.TagIDs) {
			id = row.TagIDs[i]
		}
		tags = append(tags, placemodel.EventTaxonomyItem{
			ID:   id,
			Name: name,
			Slug: slugify(name),
		})
	}

	var nextSession *placemodel.EventCardNextSession
	if row.StartAt != nil {
		nextSession = &placemodel.EventCardNextSession{
			StartAt: row.StartAt.UTC(),
			Place: placemodel.EventCardNextSessionPlace{
				Name:        row.PlaceName,
				AddressLine: row.AddressLine,
			},
		}
	}

	return placemodel.EventCardView{
		ID:               row.ID,
		Title:            row.Title,
		ShortDescription: row.ShortDesc,
		CoverImageURL:    row.CoverImageURL,
		Tags:             tags,
		NextSession:      nextSession,
	}
}

func (r *PostgresRepository) fetchCollectionEvents(ctx context.Context, collectionID string) ([]placemodel.EventCardView, error) {
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
				ARRAY_AGG(t.id::text ORDER BY t.name, t.id) AS tag_ids,
				ARRAY_AGG(t.name ORDER BY t.name, t.id) AS tag_names
			FROM event_tag et
			JOIN tag t ON t.id = et.tag_id
			GROUP BY et.event_id
		)
		SELECT
			e.id::text,
			e.title,
			e.location_description,
			COALESCE(fi.image_url, '') AS cover_image_url,
			COALESCE(tags.tag_ids, '{}'::text[]) AS tag_ids,
			COALESCE(tags.tag_names, '{}'::text[]) AS tag_names,
			ns.start_at,
			COALESCE(ns.place_name, '') AS place_name,
			COALESCE(ns.address_line, '') AS address_line,
			0 AS total_count
		FROM collection_event ce
		JOIN event e ON e.id = ce.event_id
		LEFT JOIN first_image fi ON fi.event_id = e.id
		LEFT JOIN next_session ns ON ns.event_id = e.id
		LEFT JOIN tags ON tags.event_id = e.id
		WHERE ce.collection_id = $1
		ORDER BY ce.created_at ASC, e.id ASC
	`

	rows, err := r.pool.Query(ctx, query, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]placemodel.EventCardView, 0)
	for rows.Next() {
		var row eventListRow
		if err := rows.Scan(&row.ID, &row.Title, &row.ShortDesc, &row.CoverImageURL, &row.TagIDs, &row.TagNames, &row.StartAt, &row.PlaceName, &row.AddressLine, &row.TotalCount); err != nil {
			return nil, err
		}
		items = append(items, toEventCard(row))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return items, nil
}

func toHomeFeaturedEvent(row homeFeaturedEventRow) placemodel.HomeFeaturedEvent {
	tags := make([]placemodel.HomeTag, 0, len(row.TagNames))
	for i, name := range row.TagNames {
		id := ""
		if i < len(row.TagIDs) {
			id = row.TagIDs[i]
		}
		tags = append(tags, placemodel.HomeTag{ID: id, Name: name, Slug: slugify(name)})
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

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
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
