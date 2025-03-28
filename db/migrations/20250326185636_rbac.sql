-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS permissions (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id ulid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id ulid NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id ulid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id ulid NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (user_id, role_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;

DROP TABLE IF EXISTS role_permissions;

DROP TABLE IF EXISTS permissions;

DROP TABLE IF EXISTS roles;

-- +goose StatementEnd
