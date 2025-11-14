-- Product Matching Engine for ERP Integration
-- Implements 3-layer matching strategy: Exact → Normalized → Fuzzy

-- 1. Remove external_id from products (now handled by store_product_mappings)
ALTER TABLE products DROP COLUMN IF EXISTS external_id;

-- 2. Enable fuzzy matching extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 3. Add normalized search columns to products for faster matching
ALTER TABLE products ADD COLUMN IF NOT EXISTS normalized_name TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS extracted_volume_ml DECIMAL(10, 3);
ALTER TABLE products ADD COLUMN IF NOT EXISTS extracted_weight_g DECIMAL(10, 3);

-- 4. Create index for fuzzy text search (trigram)
CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_normalized_name_trgm ON products USING gin(normalized_name gin_trgm_ops);

-- 5. Create indexes for exact matching
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_ean ON products(ean) WHERE ean IS NOT NULL;

-- 6. Function to normalize product names
CREATE OR REPLACE FUNCTION normalize_product_name(product_name TEXT)
RETURNS TEXT AS $$
DECLARE
    normalized TEXT;
BEGIN
    normalized := LOWER(product_name);
    
    -- Remove punctuation and special characters
    normalized := REGEXP_REPLACE(normalized, '[^a-z0-9\s]', ' ', 'g');
    
    -- Remove common filler words
    normalized := REGEXP_REPLACE(normalized, '\y(soft|drink|bottle|pack|packet|box|can|tin|jar|pouch)\y', '', 'g');
    
    -- Normalize units
    normalized := REGEXP_REPLACE(normalized, '\y(litre|liter|ltr|lt)\y', 'l', 'g');
    normalized := REGEXP_REPLACE(normalized, '\y(millilitre|milliliter|milli)\y', 'ml', 'g');
    normalized := REGEXP_REPLACE(normalized, '\y(kilogram|kilo)\y', 'kg', 'g');
    normalized := REGEXP_REPLACE(normalized, '\y(gram|gm)\y', 'g', 'g');
    
    -- Remove extra spaces
    normalized := REGEXP_REPLACE(normalized, '\s+', ' ', 'g');
    normalized := TRIM(normalized);
    
    RETURN normalized;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 7. Function to extract volume in ml
CREATE OR REPLACE FUNCTION extract_volume_ml(product_name TEXT)
RETURNS DECIMAL(10, 3) AS $$
DECLARE
    volume DECIMAL(10, 3);
    matches TEXT[];
BEGIN
    -- Extract patterns like "1L", "1 L", "1000ml", "1000 ml", "1 litre"
    
    -- Check for liters
    matches := REGEXP_MATCH(product_name, '(\d+\.?\d*)\s*(l|ltr|lt|litre|liter)\y', 'i');
    IF matches IS NOT NULL THEN
        volume := matches[1]::DECIMAL * 1000;
        RETURN volume;
    END IF;
    
    -- Check for milliliters
    matches := REGEXP_MATCH(product_name, '(\d+\.?\d*)\s*(ml|millilitre|milliliter)\y', 'i');
    IF matches IS NOT NULL THEN
        RETURN matches[1]::DECIMAL;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 8. Function to extract weight in grams
CREATE OR REPLACE FUNCTION extract_weight_g(product_name TEXT)
RETURNS DECIMAL(10, 3) AS $$
DECLARE
    weight DECIMAL(10, 3);
    matches TEXT[];
BEGIN
    -- Extract patterns like "1kg", "1 kg", "500g", "500 gm"
    
    -- Check for kilograms
    matches := REGEXP_MATCH(product_name, '(\d+\.?\d*)\s*(kg|kilo|kilogram)\y', 'i');
    IF matches IS NOT NULL THEN
        weight := matches[1]::DECIMAL * 1000;
        RETURN weight;
    END IF;
    
    -- Check for grams
    matches := REGEXP_MATCH(product_name, '(\d+\.?\d*)\s*(g|gm|gram)\y', 'i');
    IF matches IS NOT NULL THEN
        RETURN matches[1]::DECIMAL;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 9. Function to find matching product (3-layer strategy)
CREATE OR REPLACE FUNCTION find_matching_product(
    p_name TEXT,
    p_barcode TEXT DEFAULT NULL,
    p_sku TEXT DEFAULT NULL,
    p_ean TEXT DEFAULT NULL,
    p_store_id UUID DEFAULT NULL,
    p_external_product_id TEXT DEFAULT NULL
)
RETURNS TABLE(
    product_id UUID,
    match_type TEXT,
    confidence DECIMAL(5, 2)
) AS $$
DECLARE
    v_normalized_name TEXT;
    v_volume_ml DECIMAL(10, 3);
    v_weight_g DECIMAL(10, 3);
BEGIN
    -- LAYER 1: EXACT MATCH (Strong signals)
    
    -- Check if external_id already exists in store_products for this store
    IF p_store_id IS NOT NULL AND p_external_product_id IS NOT NULL THEN
        RETURN QUERY
        SELECT 
            sp.product_id,
            'existing_external_id'::TEXT,
            100.00::DECIMAL(5, 2)
        FROM store_products sp
        WHERE sp.store_id = p_store_id
          AND sp.external_id = p_external_product_id
          AND sp.is_available = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Match by barcode (highest confidence)
    IF p_barcode IS NOT NULL AND p_barcode != '' THEN
        RETURN QUERY
        SELECT 
            p.id,
            'barcode'::TEXT,
            100.00::DECIMAL(5, 2)
        FROM products p
        WHERE p.barcode = p_barcode
          AND p.is_active = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Match by EAN
    IF p_ean IS NOT NULL AND p_ean != '' THEN
        RETURN QUERY
        SELECT 
            p.id,
            'ean'::TEXT,
            100.00::DECIMAL(5, 2)
        FROM products p
        WHERE p.ean = p_ean
          AND p.is_active = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Match by SKU
    IF p_sku IS NOT NULL AND p_sku != '' THEN
        RETURN QUERY
        SELECT 
            p.id,
            'sku'::TEXT,
            98.00::DECIMAL(5, 2)
        FROM products p
        WHERE p.sku = p_sku
          AND p.is_active = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- LAYER 2: NORMALIZED MATCH (Medium signals)
    
    v_normalized_name := normalize_product_name(p_name);
    v_volume_ml := extract_volume_ml(p_name);
    v_weight_g := extract_weight_g(p_name);
    
    -- Match by normalized name + volume
    IF v_volume_ml IS NOT NULL THEN
        RETURN QUERY
        SELECT 
            p.id,
            'normalized_name_volume'::TEXT,
            95.00::DECIMAL(5, 2)
        FROM products p
        WHERE normalize_product_name(p.name) = v_normalized_name
          AND ABS(COALESCE(p.extracted_volume_ml, extract_volume_ml(p.name)) - v_volume_ml) < 10
          AND p.is_active = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Match by normalized name + weight
    IF v_weight_g IS NOT NULL THEN
        RETURN QUERY
        SELECT 
            p.id,
            'normalized_name_weight'::TEXT,
            95.00::DECIMAL(5, 2)
        FROM products p
        WHERE normalize_product_name(p.name) = v_normalized_name
          AND ABS(COALESCE(p.extracted_weight_g, extract_weight_g(p.name)) - v_weight_g) < 10
          AND p.is_active = true
        LIMIT 1;
        
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- LAYER 3: FUZZY MATCH (Weak signals)
    
    -- Use trigram similarity for fuzzy matching
    RETURN QUERY
    SELECT 
        p.id,
        'fuzzy'::TEXT,
        (similarity(p.name, p_name) * 100)::DECIMAL(5, 2)
    FROM products p
    WHERE similarity(p.name, p_name) > 0.45
      AND p.is_active = true
    ORDER BY similarity(p.name, p_name) DESC
    LIMIT 1;
    
    -- If no match found, return NULL (caller should create new product)
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 10. Trigger to auto-populate normalized fields on insert/update
CREATE OR REPLACE FUNCTION update_product_normalized_fields()
RETURNS TRIGGER AS $$
BEGIN
    NEW.normalized_name := normalize_product_name(NEW.name);
    NEW.extracted_volume_ml := extract_volume_ml(NEW.name);
    NEW.extracted_weight_g := extract_weight_g(NEW.name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_normalized_fields
    BEFORE INSERT OR UPDATE OF name ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_product_normalized_fields();

-- 11. Backfill normalized fields for existing products
UPDATE products 
SET normalized_name = normalize_product_name(name),
    extracted_volume_ml = extract_volume_ml(name),
    extracted_weight_g = extract_weight_g(name)
WHERE normalized_name IS NULL;

-- Add comments
COMMENT ON FUNCTION find_matching_product IS 'Three-layer product matching: 1) Exact (barcode/SKU/EAN), 2) Normalized (name+size), 3) Fuzzy (similarity). Returns product_id, match_type, and confidence score.';
COMMENT ON FUNCTION normalize_product_name IS 'Normalizes product names by removing punctuation, filler words, and standardizing units for better matching.';
COMMENT ON FUNCTION extract_volume_ml IS 'Extracts volume in milliliters from product name (e.g., "1L" → 1000, "500ml" → 500).';
COMMENT ON FUNCTION extract_weight_g IS 'Extracts weight in grams from product name (e.g., "1kg" → 1000, "500g" → 500).';
