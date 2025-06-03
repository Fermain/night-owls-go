-- +migrate Up
ALTER TABLE broadcasts ADD COLUMN title TEXT NOT NULL DEFAULT 'Alert'; 