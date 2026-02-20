-- +goose Up
CREATE TABLE events (
                        event_id     BIGSERIAL PRIMARY KEY,
                        user_id      BIGINT NOT NULL,
                        date         TIMESTAMPTZ NOT NULL,
                        is_archived  BOOLEAN NOT NULL DEFAULT FALSE,
                        description  TEXT NOT NULL DEFAULT '',
                        created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()

);