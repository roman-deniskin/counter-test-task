-- +goose Up
ALTER TABLE banner_clicks
    ALTER COLUMN banner_id TYPE BIGINT,
    ALTER COLUMN count TYPE BIGINT;

-- +goose Down
ALTER TABLE banner_clicks
    ALTER COLUMN banner_id TYPE INTEGER,
    ALTER COLUMN count TYPE INTEGER;