-- +goose Up

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id),
    name TEXT NOT NULL CHECK (btrim(name) <> ''),
    archived BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT categories_id_company_unique UNIQUE (id, company_id)
);

CREATE UNIQUE INDEX categories_company_name_active_unique
ON categories(company_id, lower(name)) WHERE archived = false;

ALTER TABLE products
ADD COLUMN category_id UUID NULL,
ADD CONSTRAINT products_category_company_fk
    FOREIGN KEY (category_id, company_id)
    REFERENCES categories(id, company_id);

-- +goose Down

ALTER TABLE products
DROP CONSTRAINT products_category_company_fk,
DROP COLUMN category_id;

DROP TABLE categories;
