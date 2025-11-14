-- Migration: Update product_variations to reference store_products instead of products
-- This makes variations store-specific, allowing different stock levels per store

-- 1. Add store_product_id column
ALTER TABLE product_variations 
ADD COLUMN IF NOT EXISTS store_product_id UUID;

-- 2. Populate store_product_id from existing product_id
-- For each variation, find the corresponding store_product
UPDATE product_variations pv
SET store_product_id = sp.id
FROM store_products sp
WHERE sp.product_id = pv.product_id
  AND pv.store_product_id IS NULL;

-- 3. Add foreign key constraint
ALTER TABLE product_variations
ADD CONSTRAINT product_variations_store_product_id_fkey 
FOREIGN KEY (store_product_id) 
REFERENCES store_products(id) 
ON DELETE CASCADE;

-- 4. Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_variations_store_product_id 
ON product_variations(store_product_id);

-- 5. Drop old product_id foreign key constraint
ALTER TABLE product_variations 
DROP CONSTRAINT IF EXISTS product_variations_product_id_fkey;

-- 6. Keep product_id column for backward compatibility but make it nullable
ALTER TABLE product_variations 
ALTER COLUMN product_id DROP NOT NULL;

-- 7. Update unique constraint to use store_product_id instead of product_id
ALTER TABLE product_variations 
DROP CONSTRAINT IF EXISTS product_variations_product_id_name_key;

-- Drop old partial index if exists
DROP INDEX IF EXISTS idx_variations_store_product_name;

-- Create unique index without WHERE clause (required for ON CONFLICT)
CREATE UNIQUE INDEX idx_variations_store_product_name 
ON product_variations(store_product_id, name);

-- 8. Add comments
COMMENT ON COLUMN product_variations.store_product_id IS 'References store_products - variations are store-specific';
COMMENT ON COLUMN product_variations.product_id IS 'Deprecated - kept for backward compatibility, use store_product_id instead';

-- Migration completed
DO $$
BEGIN
    RAISE NOTICE 'Migration completed: product_variations now references store_products';
END $$;
