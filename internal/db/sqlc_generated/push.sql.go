// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: push.sql

package db

import (
	"context"
	"database/sql"
)

const deleteSubscription = `-- name: DeleteSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = ? AND user_id = ?
`

type DeleteSubscriptionParams struct {
	Endpoint string `json:"endpoint"`
	UserID   int64  `json:"user_id"`
}

func (q *Queries) DeleteSubscription(ctx context.Context, arg DeleteSubscriptionParams) error {
	_, err := q.db.ExecContext(ctx, deleteSubscription, arg.Endpoint, arg.UserID)
	return err
}

const getAllSubscriptions = `-- name: GetAllSubscriptions :many
SELECT user_id, endpoint, p256dh_key, auth_key FROM push_subscriptions
`

type GetAllSubscriptionsRow struct {
	UserID    int64  `json:"user_id"`
	Endpoint  string `json:"endpoint"`
	P256dhKey string `json:"p256dh_key"`
	AuthKey   string `json:"auth_key"`
}

func (q *Queries) GetAllSubscriptions(ctx context.Context) ([]GetAllSubscriptionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllSubscriptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllSubscriptionsRow{}
	for rows.Next() {
		var i GetAllSubscriptionsRow
		if err := rows.Scan(
			&i.UserID,
			&i.Endpoint,
			&i.P256dhKey,
			&i.AuthKey,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSubscriptionsByUser = `-- name: GetSubscriptionsByUser :many
SELECT endpoint, p256dh_key, auth_key FROM push_subscriptions WHERE user_id = ?
`

type GetSubscriptionsByUserRow struct {
	Endpoint  string `json:"endpoint"`
	P256dhKey string `json:"p256dh_key"`
	AuthKey   string `json:"auth_key"`
}

func (q *Queries) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]GetSubscriptionsByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getSubscriptionsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetSubscriptionsByUserRow{}
	for rows.Next() {
		var i GetSubscriptionsByUserRow
		if err := rows.Scan(&i.Endpoint, &i.P256dhKey, &i.AuthKey); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertSubscription = `-- name: UpsertSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key, user_agent, platform)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(endpoint) DO UPDATE
SET p256dh_key = excluded.p256dh_key,
    auth_key   = excluded.auth_key,
    user_agent = excluded.user_agent,
    platform   = excluded.platform
`

type UpsertSubscriptionParams struct {
	UserID    int64          `json:"user_id"`
	Endpoint  string         `json:"endpoint"`
	P256dhKey string         `json:"p256dh_key"`
	AuthKey   string         `json:"auth_key"`
	UserAgent sql.NullString `json:"user_agent"`
	Platform  sql.NullString `json:"platform"`
}

func (q *Queries) UpsertSubscription(ctx context.Context, arg UpsertSubscriptionParams) error {
	_, err := q.db.ExecContext(ctx, upsertSubscription,
		arg.UserID,
		arg.Endpoint,
		arg.P256dhKey,
		arg.AuthKey,
		arg.UserAgent,
		arg.Platform,
	)
	return err
}
