-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email citext NOT NULL UNIQUE,
    password_hash text,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc')
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
