-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, first_name, last_name, email, password_hash)
    VALUES ('01JQEG0PHECS7VVSSMRWXGBTEA', 'admin', 'admin', 'admin@admin.com', '$argon2id$v=19$m=65536,t=3,p=2$TpUso0hwEOAaKN3E0I/aIA$u4EIkRmU5i957HIB8xiKzcNiwzymi35ajP9HtFvjavE')
ON CONFLICT (email)
    DO NOTHING;

INSERT INTO roles (name, description)
    VALUES ('admin', 'Administrador com acesso completo ao sistema'),
    ('manager', 'Gerente com acesso a funcionalidades administrativas limitadas'),
    ('user', 'Usuário padrão do sistema')
ON CONFLICT (name)
    DO NOTHING;

INSERT INTO permissions (name, description)
    VALUES ('users:read', 'Visualizar usuários'),
    ('users:create', 'Criar usuários'),
    ('users:update', 'Editar usuários'),
    ('users:delete', 'Excluir usuários'),
    ('roles:read', 'Visualizar roles'),
    ('roles:create', 'Criar roles'),
    ('roles:update', 'Editar roles'),
    ('roles:delete', 'Excluir roles'),
    ('permissions:read', 'Visualizar permissões'),
    ('permissions:assign', 'Atribuir permissões')
ON CONFLICT (name)
    DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'manager'
    AND p.name IN ('users:read', 'users:create', 'users:update', 'roles:read', 'permissions:read');

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'user'
    AND p.name IN ('users:read');

INSERT INTO user_roles (user_id, role_id)
SELECT
    u.id,
    r.id
FROM
    users u,
    roles r
WHERE
    u.email = 'admin@admin.com'
    AND r.name = 'admin';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM user_roles
WHERE user_id IN (
        SELECT
            id
        FROM
            users
        WHERE
            email = 'admin@admin.com')
    AND role_id IN (
        SELECT
            id
        FROM
            roles
        WHERE
            name = 'admin');

DELETE FROM role_permissions;

DELETE FROM permissions
WHERE name IN ('users:read', 'users:create', 'users:update', 'users:delete', 'roles:read', 'roles:create', 'roles:update', 'roles:delete', 'permissions:read', 'permissions:assign');

DELETE FROM roles
WHERE name IN ('admin', 'manager', 'user');

-- +goose StatementEnd
