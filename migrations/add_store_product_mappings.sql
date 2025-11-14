-- Add store_product_mappings table for ERP integration
-- This table maps external ERP product identifiers to internal product IDs
-- Critical for multi-store operations where each store's ERP uses different product IDs

CREATE TABLE IF NOT EXISTS store_product_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    external_product_id VARCHAR(255) NOT NULL, -- ERP's product ID (e.g., "ZOH-1001", "TLY-991")
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Optional ERP-specific metadata
    external_sku VARCHAR(255), -- ERP's SKU format
    external_barcode VARCHAR(255), -- ERP's barcode
    external_name TEXT, -- ERP's product name (may differ from internal)
    
    -- Sync metadata
    last_synced_at TIMESTAMP WITH TIME ZONE,
    sync_source VARCHAR(100), -- e.g., "Zoho", "Tally", "SAP", "Odoo"
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one external_product_id per store
    UNIQUE(store_id, external_product_id)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_store_product_mappings_store_id ON store_product_mappings(store_id);
CREATE INDEX IF NOT EXISTS idx_store_product_mappings_product_id ON store_product_mappings(product_id);
CREATE INDEX IF NOT EXISTS idx_store_product_mappings_external_id ON store_product_mappings(external_product_id);
CREATE INDEX IF NOT EXISTS idx_store_product_mappings_sync_source ON store_product_mappings(sync_source);
CREATE INDEX IF NOT EXISTS idx_store_product_mappings_is_active ON store_product_mappings(is_active);

-- Add trigger for updated_at
CREATE TRIGGER update_store_product_mappings_updated_at
    BEFORE UPDATE ON store_product_mappings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment
COMMENT ON TABLE store_product_mappings IS 'Maps external ERP product identifiers to internal product IDs. Each store can have multiple ERP systems with different product IDs for the same product.';
