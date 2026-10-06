-- migrations/00003_product_table.sql

-- +goose Up

CREATE TABLE Products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name TEXT NOT NULL CHECK (btrim(name) <> ''),
    sku TEXT NOT NULL CHECK (btrim(sku) <> ''),
    unit TEXT NOT NULL
        CHECK (unit IN ('piece', 'kg', 'liter', 'meter')),
    archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX products_company_sku_active_unique
ON Products(sku, company_id) where archived = false;

-- +goose Down

DROP TABLE Products