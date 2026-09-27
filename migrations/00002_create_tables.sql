-- migrations/00002_create_tables.sql

-- +goose Up

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    label text,
    code text NOT NULL UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at bigint DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY,
    name text NOT NULL,
    company_id uuid NOT NULL REFERENCES companies(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at bigint DEFAULT 0,

    UNIQUE(id, company_id)
);

CREATE TABLE role_permissions (
    user_role_id UUID NOT NULL REFERENCES user_roles(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (user_role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    first_name text NOT NULL,
    last_name text,

    phone text NOT NULL,
    email text,
    password_hash text NOT NULL,

    company_id uuid NOT NULL REFERENCES companies(id),
    user_role_id uuid NOT NULL,
    FOREIGN KEY (user_role_id, company_id)
    REFERENCES user_roles(id, company_id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at bigint DEFAULT 0
);

CREATE UNIQUE INDEX users_active_phone_unique
    ON users (phone)
    WHERE deleted_at = 0;

CREATE TABLE user_permissions (
    user_id UUID NOT NULL REFERENCES users(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (user_id, permission_id)
);

-- +goose Down
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS permissions;