-- migrations/00006_seed_permissions_v2.sql

-- +goose Up
INSERT INTO permissions(label, code)
SELECT initcap(a) || ' ' || initcap(o), o || '.' || a
FROM unnest(ARRAY['users','locations','products','categories','roles']) AS o
CROSS JOIN unnest(ARRAY['view','create','edit','archive']) AS a
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions(label, code) VALUES
    ('View Permissions',   'permissions.view'),
    ('Grant Permissions',  'permissions.grant'),
    ('Revoke Permissions', 'permissions.revoke'),
    ('Assign Roles',       'roles.assign')
ON CONFLICT (code) DO NOTHING;

-- все права существующим владельцам
INSERT INTO user_permissions(user_id, permission_id)
SELECT u.id, p.id
FROM users u
JOIN user_roles r ON r.id = u.user_role_id AND r.name = 'owner'
CROSS JOIN permissions p
WHERE u.deleted_at = 0 AND p.deleted_at = 0
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM user_permissions WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE split_part(code, '.', 1) IN ('users','locations','products','categories','roles','permissions'));
DELETE FROM permissions
WHERE split_part(code, '.', 1) IN ('users','locations','products','categories','roles','permissions');