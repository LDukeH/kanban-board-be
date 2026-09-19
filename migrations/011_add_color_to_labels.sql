-- +goose Up

ALTER TABLE labels
    ADD COLUMN color TEXT NOT NULL DEFAULT '';
