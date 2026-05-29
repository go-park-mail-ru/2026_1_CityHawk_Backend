package repository

import (
	"context"
	"database/sql"
	"errors"

	platformerrors "cityhawk/backend/internal/platform/errors"
	socialmodel "cityhawk/backend/internal/social/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *PostgresRepository) SearchInvitees(ctx context.Context, eventID, viewerID, query string, limit int) ([]socialmodel.InviteeCandidate, error) {
	if ok, err := r.eventExists(ctx, eventID); err != nil {
		return nil, err
	} else if !ok {
		return nil, platformerrors.ErrInvalidReference
	}

	rows, err := r.pool.Query(ctx, `
		WITH latest_invitation AS (
			SELECT DISTINCT ON (ei.recipient_user_id)
				ei.recipient_user_id,
				ei.status
			FROM event_invitation ei
			JOIN event_invitation_event eie ON eie.invitation_id = ei.id
			WHERE eie.event_id::text = $1
			ORDER BY ei.recipient_user_id, ei.created_at DESC, ei.id DESC
		)
		SELECT
			u.id::text,
			COALESCE(NULLIF(btrim(u.username), ''), u.email) AS username,
			u.avatar_url,
			c.id::text,
			c.name,
			c.country_name,
			c.timezone,
			EXISTS (
				SELECT 1
				FROM user_follow viewer_follow
				WHERE viewer_follow.follower_user_id::text = $2
					AND viewer_follow.followed_user_id = u.id
			) AS is_following,
			true AS is_friend,
			li.status
		FROM user_account u
		LEFT JOIN city c ON c.id = u.city_id
		LEFT JOIN latest_invitation li ON li.recipient_user_id = u.id
		JOIN user_follow eligible_follow
			ON eligible_follow.follower_user_id = u.id
			AND eligible_follow.followed_user_id::text = $2
		WHERE u.id::text <> $2
			AND NOT EXISTS (
				SELECT 1 FROM event e
				WHERE e.id::text = $1 AND e.author_user_id = u.id
			)
			AND li.recipient_user_id IS NULL
			AND (
				lower(u.username) LIKE '%' || lower($3) || '%'
				OR lower(u.email) LIKE '%' || lower($3) || '%'
			)
		ORDER BY COALESCE(NULLIF(btrim(u.username), ''), u.email) ASC, u.id ASC
		LIMIT $4
	`, eventID, viewerID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInviteeCandidates(rows)
}

func scanInviteeCandidates(rows pgx.Rows) ([]socialmodel.InviteeCandidate, error) {
	items := make([]socialmodel.InviteeCandidate, 0)
	for rows.Next() {
		var item socialmodel.InviteeCandidate
		var avatarURL, cityID, cityName, countryName, timezone, status sql.NullString
		if err := rows.Scan(&item.ID, &item.Username, &avatarURL, &cityID, &cityName, &countryName, &timezone, &item.IsFollowing, &item.IsFriend, &status); err != nil {
			return nil, err
		}
		if avatarURL.Valid {
			value := avatarURL.String
			item.AvatarURL = &value
		}
		if cityID.Valid {
			item.City = &socialmodel.City{ID: cityID.String, Name: cityName.String, CountryName: countryName.String, Timezone: timezone.String}
		}
		if status.Valid {
			value := status.String
			item.InvitationStatus = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListInvitees(ctx context.Context, eventID string) ([]socialmodel.InviteeCandidate, error) {
	if ok, err := r.eventExists(ctx, eventID); err != nil {
		return nil, err
	} else if !ok {
		return nil, platformerrors.ErrInvalidReference
	}

	rows, err := r.pool.Query(ctx, `
		WITH latest_invitation AS (
			SELECT DISTINCT ON (ei.recipient_user_id)
				ei.recipient_user_id,
				ei.status
			FROM event_invitation ei
			JOIN event_invitation_event eie ON eie.invitation_id = ei.id
			WHERE eie.event_id::text = $1
			ORDER BY ei.recipient_user_id, ei.created_at DESC, ei.id DESC
		)
		SELECT
			u.id::text,
			COALESCE(NULLIF(btrim(u.username), ''), u.email) AS username,
			u.avatar_url,
			c.id::text,
			c.name,
			c.country_name,
			c.timezone,
			false AS is_following,
			true AS is_friend,
			li.status
		FROM latest_invitation li
		JOIN user_account u ON u.id = li.recipient_user_id
		LEFT JOIN city c ON c.id = u.city_id
		ORDER BY COALESCE(NULLIF(btrim(u.username), ''), u.email) ASC, u.id ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInviteeCandidates(rows)
}

func (r *PostgresRepository) CreateInvitations(ctx context.Context, senderID, eventID string, recipientIDs []string, message string, eventSessionID *string) ([]socialmodel.Invitation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var authorID string
	if err := tx.QueryRow(ctx, `SELECT author_user_id::text FROM event WHERE id::text = $1`, eventID).Scan(&authorID); errors.Is(err, pgx.ErrNoRows) {
		return nil, platformerrors.ErrInvalidReference
	} else if err != nil {
		return nil, err
	}
	if eventSessionID != nil {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM event_session WHERE id::text = $1 AND event_id::text = $2)`, *eventSessionID, eventID).Scan(&exists); err != nil {
			return nil, err
		} else if !exists {
			return nil, platformerrors.ErrInvalidReference
		}
	}
	if err := validateInviteRecipientsAreFriends(ctx, tx, senderID, recipientIDs); err != nil {
		return nil, err
	}

	items := make([]socialmodel.Invitation, 0, len(recipientIDs))
	for _, recipientID := range recipientIDs {
		if recipientID == senderID || recipientID == authorID {
			return nil, platformerrors.ErrInvalidReference
		}
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM event_invitation ei
				JOIN event_invitation_event eie ON eie.invitation_id = ei.id
				WHERE eie.event_id::text = $1
					AND ei.recipient_user_id::text = $2
					AND ei.status = 'pending'
			)
		`, eventID, recipientID).Scan(&exists); err != nil {
			return nil, err
		} else if exists {
			return nil, platformerrors.ErrAlreadyExists
		}

		var item socialmodel.Invitation
		if err := tx.QueryRow(ctx, `
			INSERT INTO event_invitation (sender_user_id, recipient_user_id, status, message_text)
			VALUES ($1, $2, 'pending', NULLIF($3, ''))
			RETURNING id::text, sender_user_id::text, recipient_user_id::text, status, COALESCE(message_text, ''), created_at, updated_at
		`, senderID, recipientID, message).Scan(&item.ID, &item.SenderID, &item.RecipientID, &item.Status, &item.Message, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, mapPgError(err)
		}
		item.EventID = eventID
		item.EventSessionID = eventSessionID
		if _, err := tx.Exec(ctx, `INSERT INTO event_invitation_event (invitation_id, event_id) VALUES ($1, $2)`, item.ID, eventID); err != nil {
			return nil, mapPgError(err)
		}
		if eventSessionID != nil {
			if _, err := tx.Exec(ctx, `INSERT INTO event_invitation_session (invitation_id, event_session_id) VALUES ($1, $2)`, item.ID, *eventSessionID); err != nil {
				return nil, mapPgError(err)
			}
		}
		if err := createNotification(ctx, tx, recipientID, "event_invitation", senderID, eventID, eventSessionID, item.ID, nil); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func validateInviteRecipientsAreFriends(ctx context.Context, tx pgx.Tx, senderID string, recipientIDs []string) error {
	if len(recipientIDs) == 0 {
		return nil
	}

	var eligibleCount int
	if err := tx.QueryRow(ctx, `
		SELECT count(DISTINCT uf.follower_user_id)
		FROM user_follow uf
		WHERE uf.followed_user_id::text = $1
			AND uf.follower_user_id::text = ANY($2::text[])
	`, senderID, recipientIDs).Scan(&eligibleCount); err != nil {
		return err
	}
	if eligibleCount != len(recipientIDs) {
		return platformerrors.ErrOnlyFriendsInvite
	}
	return nil
}

func (r *PostgresRepository) UpdateInvitationStatus(ctx context.Context, userID, invitationID, status string) (socialmodel.Invitation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return socialmodel.Invitation{}, err
	}
	defer tx.Rollback(ctx)

	var item socialmodel.Invitation
	var sessionID sql.NullString
	var respondedAt sql.NullTime
	err = tx.QueryRow(ctx, `
		SELECT
			ei.id::text,
			eie.event_id::text,
			eis.event_session_id::text,
			ei.sender_user_id::text,
			ei.recipient_user_id::text,
			ei.status,
			COALESCE(ei.message_text, ''),
			ei.responded_at,
			ei.created_at,
			ei.updated_at
		FROM event_invitation ei
		JOIN event_invitation_event eie ON eie.invitation_id = ei.id
		LEFT JOIN event_invitation_session eis ON eis.invitation_id = ei.id
		WHERE ei.id::text = $1
	`, invitationID).Scan(&item.ID, &item.EventID, &sessionID, &item.SenderID, &item.RecipientID, &item.Status, &item.Message, &respondedAt, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return socialmodel.Invitation{}, platformerrors.ErrNotFound
	}
	if err != nil {
		return socialmodel.Invitation{}, err
	}
	if item.RecipientID != userID {
		return socialmodel.Invitation{}, platformerrors.ErrForbidden
	}
	if item.Status != "pending" {
		return socialmodel.Invitation{}, platformerrors.ErrAlreadyExists
	}
	if sessionID.Valid {
		value := sessionID.String
		item.EventSessionID = &value
	}

	err = tx.QueryRow(ctx, `
		UPDATE event_invitation
		SET status = $2, responded_at = now()
		WHERE id::text = $1
		RETURNING status, responded_at, updated_at
	`, invitationID, status).Scan(&item.Status, &respondedAt, &item.UpdatedAt)
	if err != nil {
		return socialmodel.Invitation{}, err
	}
	if respondedAt.Valid {
		value := respondedAt.Time.UTC()
		item.RespondedAt = &value
	}
	switch status {
	case "accepted", "declined":
		notificationType := "invitation_" + status
		if err := createNotification(ctx, tx, item.SenderID, notificationType, userID, item.EventID, item.EventSessionID, item.ID, nil); err != nil {
			return socialmodel.Invitation{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return socialmodel.Invitation{}, err
	}
	return item, nil
}

func (r *PostgresRepository) ListNotifications(ctx context.Context, userID, filterType string, unreadOnly bool, limit, offset int) ([]socialmodel.Notification, int, int, error) {
	typeWhere := ``
	switch filterType {
	case "invitations":
		typeWhere = `AND n.notification_type IN ('event_invitation', 'invitation_accepted', 'invitation_declined')`
	case "system":
		typeWhere = `AND n.notification_type IN ('system', 'event_reminder', 'collection_shared')`
	}
	unreadWhere := ``
	if unreadOnly {
		unreadWhere = `AND n.is_read = false`
	}
	query := `
		WITH first_image AS (
			SELECT DISTINCT ON (event_id) event_id, image_url
			FROM event_image
			ORDER BY event_id, created_at ASC, id ASC
		),
		total_rows AS (
			SELECT count(*) AS total
			FROM notification n
			WHERE n.recipient_user_id::text = $1 ` + typeWhere + ` ` + unreadWhere + `
		),
		unread_rows AS (
			SELECT count(*) AS unread_count
			FROM notification n
			WHERE n.recipient_user_id::text = $1 AND n.is_read = false
		)
		SELECT
			n.id::text,
			n.notification_type,
			n.is_read,
			n.read_at,
			n.created_at,
			COALESCE(ei.message_text, ''),
			actor.id::text,
			COALESCE(NULLIF(btrim(actor.username), ''), actor.email, 'CityHawk'),
			actor.avatar_url,
			e.id::text,
			e.title,
			COALESCE(fi.image_url, ''),
			es.id::text,
			es.start_at,
			es.end_at,
			concat_ws(', ', p.name, c.name),
			ei.id::text,
			ei.status,
			col.id::text,
			col.title,
			COALESCE(ci.image_url, ''),
			(SELECT total FROM total_rows),
			(SELECT unread_count FROM unread_rows)
		FROM notification n
		LEFT JOIN notification_actor na ON na.notification_id = n.id
		LEFT JOIN user_account actor ON actor.id = na.author_user_id
		LEFT JOIN notification_event ne ON ne.notification_id = n.id
		LEFT JOIN event e ON e.id = ne.event_id
		LEFT JOIN first_image fi ON fi.event_id = e.id
		LEFT JOIN notification_event_session nes ON nes.notification_id = n.id
		LEFT JOIN event_session es ON es.id = nes.event_session_id
		LEFT JOIN place p ON p.id = es.place_id
		LEFT JOIN city c ON c.id = p.city_id
		LEFT JOIN notification_invitation ni ON ni.notification_id = n.id
		LEFT JOIN event_invitation ei ON ei.id = ni.invitation_id
		LEFT JOIN notification_collection nc ON nc.notification_id = n.id
		LEFT JOIN collection col ON col.id = nc.collection_id
		LEFT JOIN LATERAL (
			SELECT image_url
			FROM collection_image
			WHERE collection_id = col.id
			ORDER BY created_at ASC, id ASC
			LIMIT 1
		) ci ON true
		WHERE n.recipient_user_id::text = $1 ` + typeWhere + ` ` + unreadWhere + `
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	items := make([]socialmodel.Notification, 0)
	total, unread := 0, 0
	for rows.Next() {
		var item socialmodel.Notification
		var readAt, sessionStart, sessionEnd sql.NullTime
		var message, actorID, actorName, actorAvatar, eventID, eventTitle, eventImage, sessionID, placeText, invitationID, invitationStatus, collectionID, collectionTitle, collectionImage sql.NullString
		if err := rows.Scan(&item.ID, &item.Type, &item.IsRead, &readAt, &item.CreatedAt, &message, &actorID, &actorName, &actorAvatar, &eventID, &eventTitle, &eventImage, &sessionID, &sessionStart, &sessionEnd, &placeText, &invitationID, &invitationStatus, &collectionID, &collectionTitle, &collectionImage, &total, &unread); err != nil {
			return nil, 0, 0, err
		}
		if readAt.Valid {
			value := readAt.Time.UTC()
			item.ReadAt = &value
		}
		item.Message = message.String
		if actorID.Valid {
			item.Actor = &socialmodel.NotificationActor{ID: actorID.String, DisplayName: actorName.String}
			if actorAvatar.Valid {
				value := actorAvatar.String
				item.Actor.AvatarURL = &value
			}
		}
		if eventID.Valid {
			item.Event = &socialmodel.NotificationEvent{ID: eventID.String, Title: eventTitle.String, CoverImageURL: eventImage.String, PlaceText: placeText.String}
			if sessionStart.Valid {
				item.Event.DateText = sessionStart.Time.UTC().Format("2006-01-02 15:04")
			}
		}
		if sessionID.Valid && sessionStart.Valid && sessionEnd.Valid {
			item.EventSession = &socialmodel.NotificationEventSession{ID: sessionID.String, StartAt: sessionStart.Time.UTC(), EndAt: sessionEnd.Time.UTC()}
		}
		if invitationID.Valid {
			item.Invitation = &socialmodel.NotificationInvitation{ID: invitationID.String, Status: invitationStatus.String}
		}
		if collectionID.Valid {
			item.Collection = &socialmodel.NotificationCollection{ID: collectionID.String, Title: collectionTitle.String, ImageURL: collectionImage.String}
		}
		items = append(items, item)
	}
	return items, total, unread, rows.Err()
}

func (r *PostgresRepository) MarkNotificationRead(ctx context.Context, userID, notificationID string) (int, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE notification
		SET is_read = true, read_at = COALESCE(read_at, now())
		WHERE id::text = $1 AND recipient_user_id::text = $2
	`, notificationID, userID)
	if err != nil {
		return 0, err
	}
	if tag.RowsAffected() == 0 {
		return 0, platformerrors.ErrNotFound
	}
	return r.unreadNotifications(ctx, userID)
}

func (r *PostgresRepository) ListNotificationEvents(ctx context.Context, userID, status string, limit, offset int) ([]socialmodel.NotificationEventRef, int, error) {
	rows, err := r.pool.Query(ctx, `
		WITH latest_notification_events AS (
			SELECT DISTINCT ON (ne.event_id)
				ne.event_id,
				n.created_at AS last_notification_at,
				ei.id AS invitation_id,
				ei.status AS invitation_status,
				inviter.id AS inviter_id,
				COALESCE(NULLIF(btrim(inviter.username), ''), inviter.email, 'CityHawk') AS inviter_username,
				inviter.avatar_url AS inviter_avatar_url
			FROM notification n
			JOIN notification_event ne ON ne.notification_id = n.id
			JOIN notification_invitation ni ON ni.notification_id = n.id
			JOIN event_invitation ei ON ei.id = ni.invitation_id
			JOIN user_account inviter ON inviter.id = ei.sender_user_id
			WHERE n.recipient_user_id::text = $1
			  AND ($4 = '' OR ei.status = $4)
			ORDER BY ne.event_id, n.created_at DESC, n.id DESC
		),
		notification_events AS (
			SELECT
				event_id,
				last_notification_at,
				invitation_id,
				invitation_status,
				inviter_id,
				inviter_username,
				inviter_avatar_url
			FROM latest_notification_events
		)
		SELECT
			event_id::text,
			last_notification_at,
			invitation_id::text,
			invitation_status,
			inviter_id::text,
			inviter_username,
			inviter_avatar_url,
			count(*) OVER() AS total_count
		FROM notification_events
		ORDER BY last_notification_at DESC, event_id ASC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset, status)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]socialmodel.NotificationEventRef, 0)
	total := 0
	for rows.Next() {
		var item socialmodel.NotificationEventRef
		var invitationID, invitationStatus, inviterID, inviterUsername, inviterAvatarURL sql.NullString
		if err := rows.Scan(
			&item.EventID,
			&item.CreatedAt,
			&invitationID,
			&invitationStatus,
			&inviterID,
			&inviterUsername,
			&inviterAvatarURL,
			&total,
		); err != nil {
			return nil, 0, err
		}
		if invitationID.Valid {
			item.Invitation = &socialmodel.NotificationInvitation{
				ID:     invitationID.String,
				Status: invitationStatus.String,
			}
		}
		if inviterID.Valid {
			item.InvitedBy = &socialmodel.NotificationEventInviter{
				ID:       inviterID.String,
				Username: inviterUsername.String,
			}
			if inviterAvatarURL.Valid {
				value := inviterAvatarURL.String
				item.InvitedBy.AvatarURL = &value
			}
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *PostgresRepository) MarkAllNotificationsRead(ctx context.Context, userID string) (int, error) {
	if _, err := r.pool.Exec(ctx, `
		UPDATE notification
		SET is_read = true, read_at = COALESCE(read_at, now())
		WHERE recipient_user_id::text = $1 AND is_read = false
	`, userID); err != nil {
		return 0, err
	}
	return r.unreadNotifications(ctx, userID)
}

func (r *PostgresRepository) CreateEventShareLink(ctx context.Context, creatorUserID, eventID, token string) (socialmodel.ShareLink, error) {
	if ok, err := r.eventExists(ctx, eventID); err != nil {
		return socialmodel.ShareLink{}, err
	} else if !ok {
		return socialmodel.ShareLink{}, platformerrors.ErrInvalidReference
	}
	return r.createShareLink(ctx, creatorUserID, token, &eventID, nil)
}

func (r *PostgresRepository) CreateCollectionShareLink(ctx context.Context, creatorUserID, collectionID, token string) (socialmodel.ShareLink, error) {
	if ok, err := r.collectionExists(ctx, collectionID); err != nil {
		return socialmodel.ShareLink{}, err
	} else if !ok {
		return socialmodel.ShareLink{}, platformerrors.ErrInvalidReference
	}
	return r.createShareLink(ctx, creatorUserID, token, nil, &collectionID)
}

func (r *PostgresRepository) ResolveShareLink(ctx context.Context, token string) (socialmodel.ShareLink, bool, error) {
	var item socialmodel.ShareLink
	var eventID, collectionID sql.NullString
	err := r.pool.QueryRow(ctx, `
		SELECT sl.id::text, sl.share_token, sle.event_id::text, slc.collection_id::text, sl.created_at
		FROM share_link sl
		LEFT JOIN share_link_event sle ON sle.share_link_id = sl.id
		LEFT JOIN share_link_collection slc ON slc.share_link_id = sl.id
		WHERE sl.share_token = $1
	`, token).Scan(&item.ID, &item.Token, &eventID, &collectionID, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return socialmodel.ShareLink{}, false, nil
	}
	if err != nil {
		return socialmodel.ShareLink{}, false, err
	}
	if eventID.Valid {
		value := eventID.String
		item.EventID = &value
	}
	if collectionID.Valid {
		value := collectionID.String
		item.CollectionID = &value
	}
	return item, true, nil
}

func (r *PostgresRepository) unreadNotifications(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM notification WHERE recipient_user_id::text = $1 AND is_read = false`, userID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) createShareLink(ctx context.Context, creatorUserID, token string, eventID, collectionID *string) (socialmodel.ShareLink, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return socialmodel.ShareLink{}, err
	}
	defer tx.Rollback(ctx)

	var item socialmodel.ShareLink
	var creator any
	if creatorUserID != "" {
		creator = creatorUserID
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO share_link (creator_user_id, share_token)
		VALUES ($1, $2)
		RETURNING id::text, share_token, created_at
	`, creator, token).Scan(&item.ID, &item.Token, &item.CreatedAt); err != nil {
		return socialmodel.ShareLink{}, mapPgError(err)
	}
	if eventID != nil {
		item.EventID = eventID
		if _, err := tx.Exec(ctx, `INSERT INTO share_link_event (share_link_id, event_id) VALUES ($1, $2)`, item.ID, *eventID); err != nil {
			return socialmodel.ShareLink{}, mapPgError(err)
		}
	}
	if collectionID != nil {
		item.CollectionID = collectionID
		if _, err := tx.Exec(ctx, `INSERT INTO share_link_collection (share_link_id, collection_id) VALUES ($1, $2)`, item.ID, *collectionID); err != nil {
			return socialmodel.ShareLink{}, mapPgError(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return socialmodel.ShareLink{}, err
	}
	return item, nil
}

func (r *PostgresRepository) collectionExists(ctx context.Context, collectionID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM collection WHERE id::text = $1)`, collectionID).Scan(&exists)
	return exists, err
}

type notificationTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func createNotification(ctx context.Context, tx notificationTx, recipientID, typ, actorID, eventID string, eventSessionID *string, invitationID string, collectionID *string) error {
	var notificationID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO notification (recipient_user_id, notification_type)
		VALUES ($1, $2)
		RETURNING id::text
	`, recipientID, typ).Scan(&notificationID); err != nil {
		return err
	}
	if actorID != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_actor (notification_id, author_user_id) VALUES ($1, $2)`, notificationID, actorID); err != nil {
			return err
		}
	}
	if eventID != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_event (notification_id, event_id) VALUES ($1, $2)`, notificationID, eventID); err != nil {
			return err
		}
	}
	if eventSessionID != nil {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_event_session (notification_id, event_session_id) VALUES ($1, $2)`, notificationID, *eventSessionID); err != nil {
			return err
		}
	}
	if invitationID != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_invitation (notification_id, invitation_id) VALUES ($1, $2)`, notificationID, invitationID); err != nil {
			return err
		}
	}
	if collectionID != nil {
		if _, err := tx.Exec(ctx, `INSERT INTO notification_collection (notification_id, collection_id) VALUES ($1, $2)`, notificationID, *collectionID); err != nil {
			return err
		}
	}
	return nil
}
