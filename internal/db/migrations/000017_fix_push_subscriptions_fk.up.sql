-- +migrate Up
-- Fix foreign key constraint in push_subscriptions table
-- 
-- ISSUE: push_subscriptions references users(id) but users table has user_id as primary key
-- IMPACT: This migration preserves existing subscription data while fixing the FK constraint
-- 
-- SQLite doesn't support ALTER TABLE for FK constraints, so we use the temp table approach

-- Disable foreign key checks temporarily for the migration
PRAGMA foreign_keys = OFF;

-- Create backup table with existing data
CREATE TABLE push_subscriptions_temp AS SELECT * FROM push_subscriptions;

-- Drop the existing table with incorrect foreign key
DROP TABLE push_subscriptions;

-- Recreate the table with correct foreign key constraint
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

-- Restore data from backup (if any existed) - do this before index creation for better performance
INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh_key, auth_key, user_agent, platform, created_at)
SELECT id, user_id, endpoint, p256dh_key, auth_key, user_agent, platform, created_at 
FROM push_subscriptions_temp;

-- Add index on user_id for improved query performance (after data insertion to avoid index overhead during bulk restore)
CREATE INDEX idx_push_subscriptions_user_id ON push_subscriptions(user_id);

-- Clean up temporary table
DROP TABLE push_subscriptions_temp;

-- Re-enable foreign key checks
PRAGMA foreign_keys = ON; 