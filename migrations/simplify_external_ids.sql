-- Migration: Simplify External IDs
-- Remove store_product_mappings table and add external_id directly to store_products and taxes

-- 1. Add external_id to store_products table
ALTER TABLE store_products ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- 2. Add external_id to taxes table
ALTER TABLE taxes ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- 3. Create unique constraint for store_products external_id per store
CREATE UNIQUE INDEX IF NOT EXISTS idx_store_products_external_id 
ON store_products(store_id, external_id) 
WHERE external_id IS NOT NULL;

-- 4. Create unique constraint for taxes external_id per store
CREATE UNIQUE INDEX IF NOT EXISTS idx_taxes_external_id 
ON taxes(store_id, external_id) 
WHERE external_id IS NOT NULL;

-- 5. Migrate data from store_product_mappings to store_products (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'store_product_mappings') THEN
        -- Copy external_product_id to store_products.external_id
        UPDATE store_products sp
        SET external_id = spm.external_product_id,
            updated_at = CURRENT_TIMESTAMP
        FROM store_product_mappings spm
        WHERE sp.store_id = spm.store_id 
          AND sp.product_id = spm.product_id
          AND sp.external_id IS NULL;
        
        RAISE NOTICE 'Migrated data from store_product_mappings to store_products';
    END IF;
END $$;

-- 6. Drop store_product_mappings table
DROP TABLE IF EXISTS store_product_mappings CASCADE;

-- 7. Add comments
COMMENT ON COLUMN store_products.external_id IS 'External ERP product identifier (e.g., ZOHO-1001, TALLY-991)';
COMMENT ON COLUMN taxes.external_id IS 'External ERP tax identifier (e.g., TAX-001, GST-5-EXT)';

-- 8. Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_store_products_external_id_lookup 
ON store_products(external_id) 
WHERE external_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_taxes_external_id_lookup 
ON taxes(external_id) 
WHERE external_id IS NOT NULL;

-- Migration completed successfully
