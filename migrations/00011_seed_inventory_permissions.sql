-- migrations/00011_seed_inventory_permissions.sql

-- +goose Up
INSERT INTO permissions(label, code) VALUES
    ('View Inventory', 'inventory.view')
ON CONFLICT (code) DO NOTHING;

-- системные роли (owner) получают все права
INSERT INTO role_permissions(user_role_id, permission_id)
SELECT r.id, p.id
FROM user_roles r
JOIN permissions p ON p.code = 'inventory.view' AND p.deleted_at = 0
WHERE r.is_system AND r.deleted_at = 0
ON CONFLICT DO NOTHING;

INSERT INTO user_permissions(user_id, permission_id)
SELECT u.id, rp.permission_id
FROM users u
JOIN role_permissions rp ON rp.user_role_id = u.user_role_id
JOIN permissions p ON p.id = rp.permission_id AND p.code = 'inventory.view'
WHERE u.deleted_at = 0
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM user_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'inventory.view');
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'inventory.view');
DELETE FROM permissions WHERE code = 'inventory.view';
