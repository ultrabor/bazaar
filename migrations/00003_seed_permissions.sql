-- migrations/00003_seed_permissions.sql

-- +goose Up

ALTER TABLE permissions
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

INSERT INTO permissions(label, code) VALUES
    ('Create Sales', 'sales.create'),
    ('View Debts', 'debts.view'),
    ('View Transactions', 'transactions.view');

-- +goose Down

DELETE FROM permissions 
where code in (
    'sales.create',
    'debts.view',
    'transactions.view'
    ); 


ALTER TABLE permissions
    ALTER COLUMN id DROP DEFAULT;