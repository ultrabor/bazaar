-- migrations/00007_system_roles.sql

-- +goose Up

ALTER TABLE user_roles
    ADD COLUMN IF NOT EXISTS is_system boolean NOT NULL DEFAULT false;

ALTER TABLE user_roles
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

UPDATE user_roles SET is_system = true WHERE name = 'owner';

CREATE UNIQUE INDEX IF NOT EXISTS user_roles_one_system_per_company
    ON user_roles (company_id)
    WHERE is_system AND deleted_at = 0;

INSERT INTO role_permissions(user_role_id, permission_id)
SELECT r.id, p.id
FROM user_roles r
CROSS JOIN permissions p
WHERE r.is_system AND p.deleted_at = 0
ON CONFLICT DO NOTHING;

INSERT INTO user_permissions(user_id, permission_id)
SELECT u.id, rp.permission_id
FROM users u
JOIN role_permissions rp ON rp.user_role_id = u.user_role_id
JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at = 0
WHERE u.deleted_at = 0
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS users_user_role_id_idx ON users (user_role_id);
CREATE INDEX IF NOT EXISTS role_permissions_permission_id_idx ON role_permissions (permission_id);
CREATE INDEX IF NOT EXISTS user_permissions_permission_id_idx ON user_permissions (permission_id);

-- +goose Down

DROP INDEX IF EXISTS user_permissions_permission_id_idx;
DROP INDEX IF EXISTS role_permissions_permission_id_idx;
DROP INDEX IF EXISTS users_user_role_id_idx;

DELETE FROM role_permissions rp
USING user_roles r
WHERE rp.user_role_id = r.id AND r.is_system;

DROP INDEX IF EXISTS user_roles_one_system_per_company;

ALTER TABLE user_roles ALTER COLUMN id DROP DEFAULT;
ALTER TABLE user_roles DROP COLUMN IF EXISTS is_system;