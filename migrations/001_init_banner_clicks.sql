-- +goose Up
CREATE TABLE IF NOT EXISTS banner_clicks (
    id SERIAL PRIMARY KEY,
    banner_id INTEGER NOT NULL,
    ts_minute TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT uniq_banner_minute UNIQUE (banner_id, ts_minute)
);

CREATE INDEX IF NOT EXISTS idx_banner_minute ON banner_clicks (banner_id, ts_minute);

-- +goose Down
DROP INDEX IF EXISTS idx_banner_minute;
DROP TABLE IF EXISTS banner_clicks;