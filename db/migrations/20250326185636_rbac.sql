-- +goose Up
-- +goose StatementBegin
-- Function to validate role scope matches assignment context
CREATE OR REPLACE FUNCTION scope_matches_assignment (role_id ulid, org_id ulid, branch_id ulid)
    RETURNS boolean
    AS $$
DECLARE
    role_scope text;
BEGIN
    SELECT
        scope INTO role_scope
    FROM
        roles
    WHERE
        id = role_id;
    RETURN ((role_scope = 'global')
        OR (role_scope = 'organization'
            AND org_id IS NOT NULL
            AND branch_id IS NULL)
        OR (role_scope = 'branch'
            AND org_id IS NOT NULL
            AND branch_id IS NOT NULL));
END;
$$
LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS organizations (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    external_company_id bigint,
    external_head_office_id bigint,
    name text NOT NULL,
    description text,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS branches (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    external_office_id bigint,
    cnpj text NOT NULL,
    organization_id ulid NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    name text,
    description text,
    is_warehouse boolean NOT NULL DEFAULT FALSE,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    UNIQUE (organization_id, cnpj)
);

CREATE TABLE IF NOT EXISTS roles (
    id ulid NOT NULL DEFAULT gen_monotonic_ulid () PRIMARY KEY,
    name text NOT NULL,
    description text,
    scope text NOT NULL CHECK (scope IN ('global', 'organization', 'branch')),
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    UNIQUE (name, scope)
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
    organization_id ulid REFERENCES organizations (id) ON DELETE CASCADE,
    branch_id ulid REFERENCES branches (id) ON DELETE CASCADE,
    org_id_key ulid GENERATED ALWAYS AS (COALESCE(organization_id, '00000000000000000000000000'::ulid)) STORED,
    branch_id_key ulid GENERATED ALWAYS AS (COALESCE(branch_id, '00000000000000000000000000'::ulid)) STORED,
    created_at timestamp(0) NOT NULL DEFAULT (now() at time zone 'utc'),
    PRIMARY KEY (user_id, role_id, org_id_key, branch_id_key),
    CONSTRAINT branch_must_belong_to_organization CHECK (branch_id IS NULL OR organization_id IS NOT NULL),
    CONSTRAINT scope_consistency CHECK (scope_matches_assignment (role_id, organization_id, branch_id))
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_roles;

DROP TABLE IF EXISTS role_permissions;

DROP TABLE IF EXISTS permissions;

DROP TABLE IF EXISTS roles;

DROP TABLE IF EXISTS branches;

DROP TABLE IF EXISTS organizations;

DROP FUNCTION IF EXISTS scope_matches_assignment;

-- +goose StatementEnd
