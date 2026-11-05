-- migrations/00010_inventory.sql

-- +goose Up

CREATE TABLE IF NOT EXISTS inventory_balances(
    company_id UUID NOT NULL REFERENCES companies(id),
    location_id UUID NOT NULL REFERENCES locations(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity NUMERIC(14,3) NOT NULL CHECK (quantity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (location_id, product_id)
);

CREATE TABLE IF NOT EXISTS inventory_movements(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- seq orders the history: created_at is equal for every row of one transaction.
    seq BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
    company_id UUID NOT NULL REFERENCES companies(id),
    location_id UUID NOT NULL REFERENCES locations(id),
    product_id UUID NOT NULL REFERENCES products(id),
    user_id UUID NOT NULL REFERENCES users(id),
    qty NUMERIC(14,3) NOT NULL CHECK (qty <> 0),
    reason TEXT NOT NULL CHECK (reason IN ('purchase','sale','return','transfer_in','transfer_out','stocktake','write_off')),
    ref_type TEXT CHECK (ref_type IN ('purchase','sale','return','transfer','stocktake')),
    ref_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT inventory_movements_ref_pair CHECK ((ref_type IS NULL) = (ref_id IS NULL))
);

CREATE INDEX IF NOT EXISTS inventory_movements_history_idx
ON inventory_movements(location_id, product_id, seq);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION inventory_movements_append_only()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'inventory_movements is append-only: % is not allowed', TG_OP;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER inventory_movements_no_update_delete
BEFORE UPDATE OR DELETE ON inventory_movements
FOR EACH ROW
EXECUTE FUNCTION inventory_movements_append_only();

CREATE TRIGGER inventory_movements_no_truncate
BEFORE TRUNCATE ON inventory_movements
FOR EACH STATEMENT
EXECUTE FUNCTION inventory_movements_append_only();

-- +goose Down

DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_balances;
DROP FUNCTION IF EXISTS inventory_movements_append_only();
