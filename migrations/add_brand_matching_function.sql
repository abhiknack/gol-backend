-- Add function to find or create brand with normalization

CREATE OR REPLACE FUNCTION find_or_create_brand(
    p_brand_name TEXT
)
RETURNS UUID AS $$
DECLARE
    v_brand_id UUID;
    v_normalized_name TEXT;
    v_slug TEXT;
BEGIN
    -- Return NULL if brand name is empty
    IF p_brand_name IS NULL OR TRIM(p_brand_name) = '' THEN
        RETURN NULL;
    END IF;
    
    -- Normalize the brand name
    v_normalized_name := normalize_product_name(p_brand_name);
    
    -- Try to find existing brand by exact name match
    SELECT id INTO v_brand_id
    FROM brands
    WHERE name = p_brand_name
    LIMIT 1;
    
    IF v_brand_id IS NOT NULL THEN
        RETURN v_brand_id;
    END IF;
    
    -- Try to find by normalized name match
    SELECT id INTO v_brand_id
    FROM brands
    WHERE normalized_name = v_normalized_name
    LIMIT 1;
    
    IF v_brand_id IS NOT NULL THEN
        RETURN v_brand_id;
    END IF;
    
    -- Create new brand if not found
    v_slug := LOWER(REGEXP_REPLACE(p_brand_name, '[^a-zA-Z0-9]+', '-', 'g'));
    v_slug := TRIM(BOTH '-' FROM v_slug);
    
    -- Handle slug conflicts by appending a number
    IF EXISTS (SELECT 1 FROM brands WHERE slug = v_slug) THEN
        v_slug := v_slug || '-' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT;
    END IF;
    
    INSERT INTO brands (name, slug, normalized_name)
    VALUES (p_brand_name, v_slug, v_normalized_name)
    RETURNING id INTO v_brand_id;
    
    RETURN v_brand_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION find_or_create_brand IS 'Finds existing brand by name or normalized name, or creates new brand if not found. Handles brand name variations across ERPs (Coca Cola, CocaCola, Coke).';

-- Example usage:
-- SELECT find_or_create_brand('Coca Cola');
-- SELECT find_or_create_brand('CocaCola');  -- Returns same brand_id as above
-- SELECT find_or_create_brand('Coke');      -- Returns same brand_id as above
