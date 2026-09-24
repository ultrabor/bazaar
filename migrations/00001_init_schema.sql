-- migrations/00001_init_schema.sql

-- +goose Up
CREATE TABLE companies (
    id UUID PRIMARY KEY,
    label TEXT NOT NULL,
    address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE locations (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL REFERENCES companies(id),
    type TEXT NOT NULL CHECK (type IN ('warehouse', 'store')),
    name TEXT NOT NULL,
    address TEXT,
    archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX locations_company_id_idx ON locations(company_id);

CREATE UNIQUE INDEX locations_unique_active_name_idx
ON locations (company_id, type, name)
WHERE archived = false;

-- +goose Down
DROP TABLE locations;
DROP TABLE companies;