// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: outbox.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const createOutboxItem = `-- name: CreateOutboxItem :one
INSERT INTO outbox (
    message_type,
    recipient,
    payload,
    user_id,
    send_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
)
RETURNING outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id, send_at
`

type CreateOutboxItemParams struct {
	MessageType string         `json:"message_type"`
	Recipient   string         `json:"recipient"`
	Payload     sql.NullString `json:"payload"`
	UserID      sql.NullInt64  `json:"user_id"`
	SendAt      time.Time      `json:"send_at"`
}

func (q *Queries) CreateOutboxItem(ctx context.Context, arg CreateOutboxItemParams) (Outbox, error) {
	row := q.db.QueryRowContext(ctx, createOutboxItem,
		arg.MessageType,
		arg.Recipient,
		arg.Payload,
		arg.UserID,
		arg.SendAt,
	)
	var i Outbox
	err := row.Scan(
		&i.OutboxID,
		&i.MessageType,
		&i.Recipient,
		&i.Payload,
		&i.Status,
		&i.CreatedAt,
		&i.SentAt,
		&i.RetryCount,
		&i.UserID,
		&i.SendAt,
	)
	return i, err
}

const getPendingOutboxItems = `-- name: GetPendingOutboxItems :many
SELECT outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id, send_at FROM outbox
WHERE status = 'pending'
  AND send_at <= CURRENT_TIMESTAMP
ORDER BY created_at ASC
LIMIT ?
`

func (q *Queries) GetPendingOutboxItems(ctx context.Context, limit int64) ([]Outbox, error) {
	rows, err := q.db.QueryContext(ctx, getPendingOutboxItems, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Outbox{}
	for rows.Next() {
		var i Outbox
		if err := rows.Scan(
			&i.OutboxID,
			&i.MessageType,
			&i.Recipient,
			&i.Payload,
			&i.Status,
			&i.CreatedAt,
			&i.SentAt,
			&i.RetryCount,
			&i.UserID,
			&i.SendAt,
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

const getRecentOutboxItemsByRecipient = `-- name: GetRecentOutboxItemsByRecipient :many

SELECT outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id, send_at FROM outbox
WHERE recipient = ?
ORDER BY created_at DESC
LIMIT ?
`

type GetRecentOutboxItemsByRecipientParams struct {
	Recipient string `json:"recipient"`
	Limit     int64  `json:"limit"`
}

// Limit to prevent processing too many at once
func (q *Queries) GetRecentOutboxItemsByRecipient(ctx context.Context, arg GetRecentOutboxItemsByRecipientParams) ([]Outbox, error) {
	rows, err := q.db.QueryContext(ctx, getRecentOutboxItemsByRecipient, arg.Recipient, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Outbox{}
	for rows.Next() {
		var i Outbox
		if err := rows.Scan(
			&i.OutboxID,
			&i.MessageType,
			&i.Recipient,
			&i.Payload,
			&i.Status,
			&i.CreatedAt,
			&i.SentAt,
			&i.RetryCount,
			&i.UserID,
			&i.SendAt,
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

const updateOutboxItemStatus = `-- name: UpdateOutboxItemStatus :one
UPDATE outbox
SET status = ?,
    sent_at = ?,
    retry_count = ?
WHERE outbox_id = ?
RETURNING outbox_id, message_type, recipient, payload, status, created_at, sent_at, retry_count, user_id, send_at
`

type UpdateOutboxItemStatusParams struct {
	Status     string        `json:"status"`
	SentAt     sql.NullTime  `json:"sent_at"`
	RetryCount sql.NullInt64 `json:"retry_count"`
	OutboxID   int64         `json:"outbox_id"`
}

func (q *Queries) UpdateOutboxItemStatus(ctx context.Context, arg UpdateOutboxItemStatusParams) (Outbox, error) {
	row := q.db.QueryRowContext(ctx, updateOutboxItemStatus,
		arg.Status,
		arg.SentAt,
		arg.RetryCount,
		arg.OutboxID,
	)
	var i Outbox
	err := row.Scan(
		&i.OutboxID,
		&i.MessageType,
		&i.Recipient,
		&i.Payload,
		&i.Status,
		&i.CreatedAt,
		&i.SentAt,
		&i.RetryCount,
		&i.UserID,
		&i.SendAt,
	)
	return i, err
}
