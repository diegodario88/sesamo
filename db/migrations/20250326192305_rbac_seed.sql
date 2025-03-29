-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, first_name, last_name, email, password_hash)
    VALUES ('01JQEG0PHECS7VVSSMRWXGBTEA', 'admin', 'admin', 'admin@admin.com', '$argon2id$v=19$m=65536,t=3,p=2$TpUso0hwEOAaKN3E0I/aIA$u4EIkRmU5i957HIB8xiKzcNiwzymi35ajP9HtFvjavE')
ON CONFLICT (email)
    DO NOTHING;

INSERT INTO organizations (id, external_company_id, external_head_office_id, name, description)
    VALUES ('01JQEYB8V8AZW0TCJFM5848NQX', 1, 10001, 'Carrara Holding', 'Empreendimentos imobiliários e contrução civil');

INSERT INTO branches (id, external_office_id, cnpj, organization_id, name, description, is_warehouse)
    VALUES ('01JQEYSXETE6CFC5F6VD40D0RW', 10001, '51482746000110', '01JQEYB8V8AZW0TCJFM5848NQX', 'Matriz', 'Sede da empresa', FALSE),
    ('01JQEYSXETE6CFC5F6VD40D0RX', 10002, '51482746000111', '01JQEYB8V8AZW0TCJFM5848NQX', 'Filial 1', 'Filial 1 depósito', TRUE),
    ('01JQEYSXETE6CFC5F6VD40D0RY', 10003, '51482746000112', '01JQEYB8V8AZW0TCJFM5848NQX', 'Filial 2', 'Filial 2 da empresa', FALSE);

INSERT INTO roles (name, description, scope)
    VALUES
        -- Global roles
        ('super_admin', 'Super Administrador com acesso total ao sistema', 'global'),
        ('system_admin', 'Administrador do Sistema', 'global'),
        -- Organization roles
        ('org_admin', 'Administrador da Organização', 'organization'),
        ('org_manager', 'Gerente da Organização', 'organization'),
        ('org_viewer', 'Visualizador da Organização', 'organization'),
        -- Branch roles
        ('branch_manager', 'Gerente da Filial', 'branch'),
        ('branch_employee', 'Funcionário da Filial', 'branch'),
        ('branch_viewer', 'Visualizador da Filial', 'branch')
    ON CONFLICT (name, scope)
        DO NOTHING;

INSERT INTO permissions (name, description)
    VALUES
        -- User permissions
        ('users:read', 'Visualizar usuários'),
        ('users:create', 'Criar usuários'),
        ('users:update', 'Editar usuários'),
        ('users:delete', 'Excluir usuários'),
        -- Role permissions
        ('roles:read', 'Visualizar perfis'),
        ('roles:create', 'Criar perfis'),
        ('roles:update', 'Editar perfis'),
        ('roles:delete', 'Excluir perfis'),
        ('roles:assign', 'Atribuir perfis'),
        -- Permission management
        ('permissions:read', 'Visualizar permissões'),
        ('permissions:assign', 'Atribuir permissões'),
        -- Organization permissions
        ('organizations:read', 'Visualizar organizações'),
        ('organizations:create', 'Criar organizações'),
        ('organizations:update', 'Editar organizações'),
        ('organizations:delete', 'Excluir organizações'),
        -- Branch permissions
        ('branches:read', 'Visualizar filiais'),
        ('branches:create', 'Criar filiais'),
        ('branches:update', 'Editar filiais'),
        ('branches:delete', 'Excluir filiais')
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
    r.name = 'super_admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'system_admin'
    AND p.name IN ('users:read', 'users:create', 'users:update', 'organizations:read', 'organizations:create', 'permissions:read');

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'org_admin'
    AND p.name IN ('users:read', 'users:create', 'users:update', 'branches:read', 'branches:create');

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'org_manager'
    AND p.name IN ('users:read', 'branches:read');

INSERT INTO role_permissions (role_id, permission_id)
SELECT
    r.id,
    p.id
FROM
    roles r,
    permissions p
WHERE
    r.name = 'branch_manager'
    AND p.name IN ('users:read', 'users:create');

INSERT INTO user_roles (user_id, role_id, organization_id, branch_id)
SELECT
    u.id,
    r.id,
    NULL,
    NULL
FROM
    users u,
    roles r
WHERE
    u.email = 'admin@admin.com'
    AND r.name = 'super_admin'
    AND r.scope = 'global';

-- Also assign admin as org_admin for Carrara Holding
INSERT INTO user_roles (user_id, role_id, organization_id, branch_id)
SELECT
    u.id,
    r.id,
    o.id,
    NULL
FROM
    users u,
    roles r,
    organizations o
WHERE
    u.email = 'admin@admin.com'
    AND r.name = 'org_admin'
    AND r.scope = 'organization'
    AND o.name = 'Carrara Holding';

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
            email = 'admin@admin.com');

DELETE FROM role_permissions;

DELETE FROM permissions
WHERE name IN ('users:read', 'users:create', 'users:update', 'users:delete', 'roles:read', 'roles:create', 'roles:update', 'roles:delete', 'permissions:read', 'permissions:assign');

DELETE FROM roles
WHERE name IN ('admin', 'manager', 'user');

-- +goose StatementEnd
