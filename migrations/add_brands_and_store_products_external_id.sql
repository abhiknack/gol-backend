-- Add brands table for normalized brand mapping across ERPs
-- Add external_id to store_products for ERP integration

-- 1. Create brands table
CREATE TABLE IF NOT EXISTS brands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    normalized_name TEXT, -- Auto-populated for matching
    description TEXT,
    logo_url TEXT,
    website_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Add brand_id to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS brand_id UUID REFERENCES brands(id) ON DELETE SET NULL;

-- 3. Add external_id to store_products
ALTER TABLE store_products ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- 4. Add unique constraint for store_products external_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'store_products_store_id_external_id_key'
    ) THEN
        ALTER TABLE store_products ADD CONSTRAINT store_products_store_id_external_id_key UNIQUE(store_id, external_id);
    END IF;
END $$;

-- 5. Add indexes
CREATE INDEX IF NOT EXISTS idx_brands_slug ON brands(slug);
CREATE INDEX IF NOT EXISTS idx_brands_normalized_name ON brands(normalized_name);
CREATE INDEX IF NOT EXISTS idx_brands_is_active ON brands(is_active);
CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_store_products_external_id ON store_products(external_id) WHERE external_id IS NOT NULL;

-- 6. Add trigger for brand normalized_name
CREATE OR REPLACE FUNCTION update_brand_normalized_name()
RETURNS TRIGGER AS $$
BEGIN
    NEW.normalized_name := normalize_product_name(NEW.name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_brand_normalized_name ON brands;
CREATE TRIGGER trigger_update_brand_normalized_name
    BEFORE INSERT OR UPDATE OF name ON brands
    FOR EACH ROW
    EXECUTE FUNCTION update_brand_normalized_name();

-- 7. Add trigger for updated_at on brands
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;
CREATE TRIGGER update_brands_updated_at
    BEFORE UPDATE ON brands
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 8. Migrate existing brand TEXT values to brands table (optional)
-- This creates brand entries from existing product.brand values
INSERT INTO brands (name, slug, normalized_name)
SELECT DISTINCT 
    brand,
    LOWER(REGEXP_REPLACE(brand, '[^a-zA-Z0-9]+', '-', 'g')) as slug,
    normalize_product_name(brand) as normalized_name
FROM products
WHERE brand IS NOT NULL 
  AND brand != ''
  AND NOT EXISTS (
      SELECT 1 FROM brands b WHERE b.name = products.brand
  )
ON CONFLICT (name) DO NOTHING;

-- 9. Update products to use brand_id
UPDATE products p
SET brand_id = b.id
FROM brands b
WHERE p.brand = b.name
  AND p.brand_id IS NULL;

-- Add comments
COMMENT ON TABLE brands IS 'Normalized brand names for multi-ERP mapping. Different ERPs may use different brand names (Coca Cola, CocaCola, Coke) which map to the same brand.';
COMMENT ON COLUMN brands.normalized_name IS 'Auto-populated normalized brand name for matching across ERPs.';
COMMENT ON COLUMN products.brand_id IS 'Foreign key to brands table. Use this instead of brand TEXT field.';
COMMENT ON COLUMN products.brand IS 'Deprecated: Use brand_id instead. Kept for backward compatibility.';
COMMENT ON COLUMN store_products.external_id IS 'ERP-specific identifier for this store-product combination.';

-- Example usage:
-- Find or create brand:
-- INSERT INTO brands (name, slug) VALUES ('Coca Cola', 'coca-cola')
-- ON CONFLICT (name) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
-- RETURNING id;
--
-- Use brand_id in products:
-- UPDATE products SET brand_id = 'brand-uuid' WHERE id = 'product-uuid';
--
-- Brand matching across ERPs:
-- SELECT * FROM brands WHERE normalized_name = normalize_product_name('CocaCola');
