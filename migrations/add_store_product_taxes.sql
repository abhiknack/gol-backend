-- Add store_product_taxes table for store-specific tax configuration
-- This handles cases where taxes vary by store due to:
-- - Different states/regions (Karnataka 18%, Delhi 12%)
-- - Store-specific GST registration (registered vs composition scheme)
-- - Local taxes, service tax, additional cess
-- - Franchise models with different tax schemes

CREATE TABLE IF NOT EXISTS store_product_taxes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    store_product_id UUID NOT NULL REFERENCES store_products(id) ON DELETE CASCADE,
    tax_id UUID NOT NULL REFERENCES taxes(id) ON DELETE CASCADE,
    
    -- Optional: Override tax rate for this specific store+product combination
    override_rate DECIMAL(5, 2), -- If NULL, use tax.rate from taxes table
    
    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(store_id, store_product_id, tax_id)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_store_product_taxes_store_id ON store_product_taxes(store_id);
CREATE INDEX IF NOT EXISTS idx_store_product_taxes_store_product_id ON store_product_taxes(store_product_id);
CREATE INDEX IF NOT EXISTS idx_store_product_taxes_tax_id ON store_product_taxes(tax_id);
CREATE INDEX IF NOT EXISTS idx_store_product_taxes_is_active ON store_product_taxes(is_active);

-- Add trigger for updated_at
CREATE TRIGGER update_store_product_taxes_updated_at
    BEFORE UPDATE ON store_product_taxes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment
COMMENT ON TABLE store_product_taxes IS 'Store-specific tax configuration. Use when taxes vary by store due to different states, GST registration, local taxes, or franchise models. If tax is same for all stores, use product_taxes instead.';
COMMENT ON COLUMN store_product_taxes.override_rate IS 'Optional tax rate override for this store+product. If NULL, uses the rate from taxes table.';

-- Example usage:
-- Store in Karnataka (18% GST):
-- INSERT INTO store_product_taxes (store_id, store_product_id, tax_id) 
-- VALUES ('karnataka-store-uuid', 'store-product-uuid', 'gst-18-tax-uuid');
--
-- Store in Delhi (12% GST):
-- INSERT INTO store_product_taxes (store_id, store_product_id, tax_id) 
-- VALUES ('delhi-store-uuid', 'store-product-uuid', 'gst-12-tax-uuid');
--
-- Store with special rate override:
-- INSERT INTO store_product_taxes (store_id, store_product_id, tax_id, override_rate) 
-- VALUES ('special-store-uuid', 'store-product-uuid', 'gst-tax-uuid', 5.00);
