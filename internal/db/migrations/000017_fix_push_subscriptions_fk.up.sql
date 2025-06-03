-- +migrate Up
-- Fix foreign key constraint in push_subscriptions table
-- The issue: push_subscriptions references users(id) but users table has user_id as primary key

-- Drop the existing table and recreate with correct foreign key
DROP TABLE IF EXISTS push_subscriptions;

CREATE TABLE push_subscriptions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER     NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    endpoint    TEXT UNIQUE NOT NULL,
    p256dh_key  TEXT        NOT NULL,
    auth_key    TEXT        NOT NULL,
    user_agent  TEXT,
    platform    TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
); 