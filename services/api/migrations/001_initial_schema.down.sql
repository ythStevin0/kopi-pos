-- migrations/001_initial_schema.down.sql
-- Rollback: hapus semua yang dibuat oleh 001_initial_schema.up.sql

DROP TRIGGER IF EXISTS set_updated_at_products   ON products;
DROP TRIGGER IF EXISTS set_updated_at_ingredients ON ingredients;
DROP FUNCTION IF EXISTS trigger_set_updated_at;

DROP TABLE IF EXISTS sale_items;
DROP TABLE IF EXISTS sales;
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS ingredients;
