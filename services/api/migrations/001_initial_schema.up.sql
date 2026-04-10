-- migrations/001_initial_schema.up.sql
-- KopiPOS: Initial Database Schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- TABLE: ingredients (Bahan baku)
-- ============================================================
CREATE TABLE ingredients (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150)   NOT NULL,
    unit          VARCHAR(30)    NOT NULL,           -- e.g., 'gram', 'ml', 'pcs'
    stock         NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (stock >= 0),
    min_stock     NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL               -- Soft Delete
);

-- Partial index: hanya baris aktif yang ter-index (performa optimal)
CREATE INDEX idx_ingredients_deleted_at ON ingredients (deleted_at)
    WHERE deleted_at IS NULL;

-- ============================================================
-- TABLE: products (Menu/Produk)
-- ============================================================
CREATE TABLE products (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(150)   NOT NULL,
    description   TEXT,
    price         NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    category      VARCHAR(80),
    image_url     TEXT,
    is_available  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ    NULL               -- Soft Delete
);

CREATE INDEX idx_products_deleted_at ON products (deleted_at)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products (category)
    WHERE deleted_at IS NULL;

-- ============================================================
-- TABLE: recipes (Junction: product ↔ ingredients)
-- ============================================================
CREATE TABLE recipes (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID           NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    ingredient_id   UUID           NOT NULL REFERENCES ingredients(id) ON DELETE RESTRICT,
    usage_quantity  NUMERIC(12, 4) NOT NULL CHECK (usage_quantity > 0),  -- Jumlah bahan per 1 produk
    UNIQUE (product_id, ingredient_id)
);

CREATE INDEX idx_recipes_product_id    ON recipes (product_id);
CREATE INDEX idx_recipes_ingredient_id ON recipes (ingredient_id);

-- ============================================================
-- TABLE: sales (Transaksi Penjualan)
-- ============================================================
CREATE TABLE sales (
    id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key  UUID           NOT NULL UNIQUE,    -- Pencegah double-payment
    cashier_id       UUID,                              -- FK ke tabel users (opsional)
    total_amount     NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
    payment_method   VARCHAR(50)    NOT NULL,           -- 'cash', 'qris', 'card'
    status           VARCHAR(30)    NOT NULL DEFAULT 'completed'
                     CHECK (status IN ('pending', 'completed', 'voided')),
    notes            TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sales_idempotency_key ON sales (idempotency_key);
CREATE INDEX idx_sales_created_at      ON sales (created_at DESC);

-- ============================================================
-- TABLE: sale_items (Detail Item per Transaksi)
-- ============================================================
CREATE TABLE sale_items (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id      UUID           NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    product_id   UUID           NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    product_name VARCHAR(150)   NOT NULL,               -- Snapshot nama (anti-perubahan retroaktif)
    quantity     INT            NOT NULL CHECK (quantity > 0),
    unit_price   NUMERIC(12, 2) NOT NULL CHECK (unit_price >= 0),
    subtotal     NUMERIC(12, 2) GENERATED ALWAYS AS (quantity * unit_price) STORED
);

CREATE INDEX idx_sale_items_sale_id    ON sale_items (sale_id);
CREATE INDEX idx_sale_items_product_id ON sale_items (product_id);

-- ============================================================
-- FUNCTION & TRIGGER: auto-update updated_at
-- ============================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at_ingredients
    BEFORE UPDATE ON ingredients
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_products
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
