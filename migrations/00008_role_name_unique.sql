-- migrations/00008_role_name_unique.sql

-- +goose Up

-- Role names are unique per company among active roles, case-insensitive.
CREATE UNIQUE INDEX IF NOT EXISTS user_roles_company_name_unique
    ON user_roles (company_id, lower(name))
    WHERE deleted_at = 0;

-- +goose Down

DROP INDEX IF EXISTS user_roles_company_name_unique;
