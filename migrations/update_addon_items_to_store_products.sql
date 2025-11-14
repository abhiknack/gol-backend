-- Update product_addon_group_items to reference store_products instead of products
-- This allows store-specific addon pricing and availability

-- 1. Drop existing foreign key constraint
ALTER TABLE product_addon_group_items 
DROP CONSTRAINT IF EXISTS product_addon_group_items_addon_product_id_fkey;

-- 2. Rename column
ALTER TABLE product_addon_group_items 
RENAME COLUMN addon_product_id TO addon_store_product_id;

-- 3. Drop old unique constraint
ALTER TABLE product_addon_group_items 
DROP CONSTRAINT IF EXISTS product_addon_group_items_addon_group_id_addon_product_id_key;

-- 4. Add new foreign key to store_products
ALTER TABLE product_addon_group_items 
ADD CONSTRAINT product_addon_group_items_addon_store_product_id_fkey 
FOREIGN KEY (addon_store_product_id) REFERENCES store_products(id) ON DELETE CASCADE;

-- 5. Add new unique constraint
ALTER TABLE product_addon_group_items 
ADD CONSTRAINT product_addon_group_items_addon_group_id_addon_store_product_id_key 
UNIQUE (addon_group_id, addon_store_product_id);

-- 6. Drop old index
DROP INDEX IF EXISTS idx_addon_group_items_product_id;

-- 7. Create new index
CREATE INDEX IF NOT EXISTS idx_addon_group_items_store_product_id 
ON product_addon_group_items(addon_store_product_id);

-- Add comment
COMMENT ON COLUMN product_addon_group_items.addon_store_product_id IS 'References store_products for store-specific addon pricing and availability. Allows different stores to have different addon prices and stock levels.';

-- Note: This is a breaking change. Existing data in product_addon_group_items
-- will need to be migrated to reference store_products instead of products.
-- 
-- Migration strategy:
-- 1. For each addon_product_id, find corresponding store_product_id
-- 2. Update the reference
-- 3. Handle cases where product doesn't exist in store_products
--
-- Example migration (if you have existing data):
-- UPDATE product_addon_group_items pagi
-- SET addon_store_product_id = sp.id
-- FROM store_products sp
-- WHERE sp.product_id = pagi.addon_store_product_id  -- old value
--   AND sp.store_id = (
--       SELECT store_id FROM ... -- determine store context
--   );
