-- Remove product_taxes table
-- We only use store_product_taxes for store-specific tax configuration
-- This provides more flexibility for multi-store operations where taxes vary by location

DROP TABLE IF EXISTS product_taxes CASCADE;

-- Note: All tax configuration is now done via store_product_taxes table
-- This allows different stores to have different tax rates for the same product
