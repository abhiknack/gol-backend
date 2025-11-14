
-- ============================================================
-- GROCERY SUPERAPP DATABASE SCHEMA
-- PostgreSQL Schema for Multi-Store Grocery Platform
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis"; -- For location-based features

-- ============================================================
-- CORE ENTITIES
-- ============================================================

-- Users/Customers Table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    profile_image_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- User Addresses
CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(50), -- 'Home', 'Work', 'Other'
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location GEOGRAPHY(POINT, 4326), -- PostGIS point
    is_default BOOLEAN DEFAULT FALSE,
    delivery_instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Stores/Shops Table
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE, -- Integration with external systems
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    store_type VARCHAR(50) NOT NULL, -- 'supermarket', 'pharmacy', 'convenience'
    logo_url TEXT,
    banner_url TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),

    -- Address & Location
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    location GEOGRAPHY(POINT, 4326), -- PostGIS point

    -- Business Details
    rating DECIMAL(3, 2) DEFAULT 0.00,
    total_ratings INTEGER DEFAULT 0,
    min_order_amount DECIMAL(10, 2) DEFAULT 0.00,
    delivery_fee DECIMAL(10, 2) DEFAULT 0.00,
    delivery_fee_currency VARCHAR(3) DEFAULT 'INR',
    estimated_delivery_time INTEGER, -- in minutes

    -- Status & Flags
    is_active BOOLEAN DEFAULT TRUE,
    is_open BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    is_sponsored BOOLEAN DEFAULT FALSE,
    accepts_cod BOOLEAN DEFAULT TRUE,
    has_in_store_prices BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    opened_at TIME,
    closed_at TIME
);

-- Store Operating Hours
CREATE TABLE store_hours (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday, 6=Saturday
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Categories Table (Hierarchical)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    icon_url TEXT,
    image_url TEXT,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Brands Table (Normalized brand names for multi-ERP mapping)
CREATE TABLE brands (
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

-- Products Table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    urn VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,

    -- Pricing
    base_price DECIMAL(10, 2) NOT NULL,
    sale_price DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'INR',

    -- Inventory
    unit VARCHAR(50), -- 'kg', 'liter', 'piece', 'pack'
    unit_quantity DECIMAL(10, 3) DEFAULT 1,
    is_weighted BOOLEAN DEFAULT FALSE,

    -- Media
    primary_image_url TEXT,
    image_id VARCHAR(100),

    -- Product Details
    brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,
    brand VARCHAR(100), -- Deprecated: Use brand_id instead (kept for backward compatibility)
    manufacturer VARCHAR(255),
    barcode VARCHAR(50),
    ean VARCHAR(50),

    -- Tax
    tax_ids TEXT, -- Comma-separated tax IDs (e.g., "uuid1,uuid2")

    -- Flags
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    is_new BOOLEAN DEFAULT FALSE,
    is_customizable BOOLEAN DEFAULT FALSE,
    is_addon BOOLEAN DEFAULT FALSE, -- Product can be used as addon
    requires_prescription BOOLEAN DEFAULT FALSE, -- For pharmacy items

    -- SEO & Meta
    meta_title VARCHAR(255),
    meta_description TEXT,
    
    -- Product Matching Engine Fields (for ERP integration)
    normalized_name TEXT, -- Auto-populated normalized name for matching
    extracted_volume_ml DECIMAL(10, 3), -- Auto-extracted volume in ml
    extracted_weight_g DECIMAL(10, 3), -- Auto-extracted weight in grams

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Images (Multiple images per product)
CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    display_order INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (product_id, image_url)
);

-- Store Products (Inventory per store)
CREATE TABLE store_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255), -- ERP's store-product identifier
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,

    -- Store-specific pricing
    price DECIMAL(10, 2) NOT NULL,
    sale_price DECIMAL(10, 2),

    -- Inventory
    stock_quantity DECIMAL(10, 3) DEFAULT 0,
    low_stock_threshold DECIMAL(10, 3) DEFAULT 10,
    is_in_stock BOOLEAN DEFAULT TRUE,

    -- Flags
    is_available BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(store_id, product_id),
    UNIQUE(store_id, external_id)
);

-- ============================================================
-- PRODUCT VARIATIONS & ADDONS
-- ============================================================

-- Tax Configuration
CREATE TABLE taxes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255), -- ERP's tax identifier (e.g., "TAX-001", "GST-5-EXT")
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- 'GST', 'VAT', 'Service Tax'
    tax_id VARCHAR(50) NOT NULL, -- 'GST_5', 'GST_12', 'GST_18'
    description TEXT,
    rate DECIMAL(5, 2) NOT NULL, -- Tax rate percentage (e.g., 5.00, 12.00, 18.00)
    tax_type VARCHAR(50) NOT NULL, -- 'percentage', 'fixed'
    is_inclusive BOOLEAN DEFAULT FALSE, -- Tax included in price or added on top
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(store_id, tax_id),
    UNIQUE(store_id, external_id)
);

-- Store Product Taxes (Store-specific tax configuration)
-- Use this when taxes vary by store due to:
-- - Different states/regions (Karnataka 18%, Delhi 12%)
-- - Store-specific GST registration (registered vs composition scheme)
-- - Local taxes, service tax, additional cess
-- - Franchise models with different tax schemes
CREATE TABLE store_product_taxes (
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

-- Product Variations (e.g., "Small 250ml", "Medium 500ml", "Large 1L")
-- Store-specific: Each store can have different stock/pricing for variations
CREATE TABLE product_variations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(100) UNIQUE, -- Integration with external systems
    store_product_id UUID NOT NULL REFERENCES store_products(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE CASCADE, -- Deprecated: kept for backward compatibility
    name VARCHAR(100) NOT NULL, -- 'Small', 'Medium', 'Large'
    display_name VARCHAR(255) NOT NULL, -- '250ml', '500ml', '1L'
    description TEXT,
    sku_suffix VARCHAR(50), -- Append to product SKU
    
    -- Pricing
    price DECIMAL(10, 2) NOT NULL, -- Variation price
    sale_price DECIMAL(10, 2), -- Variation sale price
    currency VARCHAR(3) DEFAULT 'INR',
    
    -- Inventory (store-specific)
    stock_quantity DECIMAL(10, 3),
    is_in_stock BOOLEAN DEFAULT TRUE,
    
    -- Display & Status
    display_order INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Unique constraint for upsert operations (store-specific)
    UNIQUE (store_product_id, name),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Addon Groups (e.g., "Extra Toppings", "Choose Your Sides")
-- Groups contain products (marked with is_addon=true) as addon options
CREATE TABLE product_addon_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- 'Extra Toppings', 'Sides'
    display_name VARCHAR(255) NOT NULL, -- 'Add Extra Toppings', 'Choose Your Sides'
    description TEXT,
    
    -- Selection Rules
    is_required BOOLEAN DEFAULT FALSE,
    min_selections INTEGER DEFAULT 0,
    max_selections INTEGER, -- NULL for unlimited
    
    -- Display
    display_order INTEGER DEFAULT 0,
    display_type VARCHAR(50) DEFAULT 'checkbox', -- 'checkbox', 'radio', 'quantity'
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Links store products (with is_addon=true) to addon groups
-- Uses store_products instead of products for store-specific pricing and availability
CREATE TABLE product_addon_group_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    addon_group_id UUID NOT NULL REFERENCES product_addon_groups(id) ON DELETE CASCADE,
    addon_store_product_id UUID NOT NULL REFERENCES store_products(id) ON DELETE CASCADE,
    
    -- Pricing override (optional, if NULL uses store_product's price)
    price_override DECIMAL(10, 2),
    use_sale_price BOOLEAN DEFAULT TRUE, -- Use store_product's sale_price if available
    
    -- Display & Status
    display_order INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT FALSE, -- Pre-selected by default
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(addon_group_id, addon_store_product_id)
);

-- ============================================================
-- ORDERS & TRANSACTIONS
-- ============================================================

-- Orders Table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE RESTRICT,

    -- Delivery Information
    delivery_address_id UUID REFERENCES user_addresses(id),
    delivery_address_snapshot JSONB, -- Store address at time of order
    delivery_instructions TEXT,

    -- Pricing
    subtotal DECIMAL(10, 2) NOT NULL,
    delivery_fee DECIMAL(10, 2) DEFAULT 0.00,
    tax_amount DECIMAL(10, 2) DEFAULT 0.00,
    discount_amount DECIMAL(10, 2) DEFAULT 0.00,
    total_amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending', 
    -- 'pending', 'confirmed', 'preparing', 'ready', 'out_for_delivery', 'delivered', 'cancelled'
    payment_status VARCHAR(50) DEFAULT 'pending',
    -- 'pending', 'paid', 'failed', 'refunded'

    -- Fulfillment
    order_type VARCHAR(50) NOT NULL DEFAULT 'delivery', 
    -- 'delivery', 'pickup', 'scheduled_delivery', 'scheduled_pickup'
    scheduled_time TIMESTAMP WITH TIME ZONE, -- For scheduled orders
    handling_type VARCHAR(50) DEFAULT 'delivery', -- 'delivery', 'pickup' (deprecated, use order_type)
    estimated_delivery_time TIMESTAMP WITH TIME ZONE,
    actual_delivery_time TIMESTAMP WITH TIME ZONE,

    -- Assignment
    driver_id UUID, -- Reference to delivery driver

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancellation_reason TEXT
);

-- Order Items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    store_product_id UUID REFERENCES store_products(id),

    -- Product Details (snapshot at time of order)
    product_name VARCHAR(255) NOT NULL,
    product_sku VARCHAR(100),
    product_image_url TEXT,

    -- Pricing
    unit_price DECIMAL(10, 2) NOT NULL,
    quantity DECIMAL(10, 3) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) DEFAULT 0.00,
    total_amount DECIMAL(10, 2) NOT NULL,

    -- Customization
    attributes JSONB, -- Store selected attributes (DEPRECATED)
    variations_snapshot JSONB, -- Selected variations at time of order
    addons_snapshot JSONB, -- Selected addons at time of order
    special_instructions TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Order Item Variations (Selected variation for order items)
CREATE TABLE order_item_variations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    variation_id UUID REFERENCES product_variations(id) ON DELETE SET NULL,
    
    -- Snapshot data (in case variation is deleted)
    variation_name VARCHAR(100) NOT NULL,
    variation_display_name VARCHAR(255) NOT NULL,
    variation_price DECIMAL(10, 2) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_item_id)
);

-- Order Item Addons (Selected addons for order items)
CREATE TABLE order_item_addons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    addon_group_id UUID REFERENCES product_addon_groups(id) ON DELETE SET NULL,
    addon_product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    
    -- Snapshot data (in case addon is deleted)
    addon_group_name VARCHAR(100) NOT NULL,
    addon_product_name VARCHAR(255) NOT NULL,
    addon_product_sku VARCHAR(100),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Order Status History
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    notes TEXT,
    created_by UUID, -- User or system that changed status
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- PROMOTIONS & DISCOUNTS
-- ============================================================

-- Promotions/Offers Table
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    promo_code VARCHAR(50) UNIQUE,

    -- Promotion Type
    discount_type VARCHAR(50) NOT NULL, -- 'percentage', 'fixed_amount', 'free_delivery'
    discount_value DECIMAL(10, 2),
    max_discount_amount DECIMAL(10, 2),
    min_order_amount DECIMAL(10, 2),

    -- Scope
    applies_to VARCHAR(50) NOT NULL, -- 'store', 'category', 'product', 'all'

    -- Limitations
    usage_limit INTEGER, -- Total usage limit
    usage_limit_per_user INTEGER DEFAULT 1,
    current_usage_count INTEGER DEFAULT 0,

    -- Timing
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Flags
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Promotion Stores (Which stores the promotion applies to)
CREATE TABLE promotion_stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    UNIQUE(promotion_id, store_id)
);

-- Promotion Products (Which products the promotion applies to)
CREATE TABLE promotion_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    UNIQUE(promotion_id, product_id)
);

-- Promotion Usage Tracking
CREATE TABLE promotion_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    discount_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- SEARCH & FILTERING
-- ============================================================

-- Filter Categories
CREATE TABLE filters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    filter_type VARCHAR(50) NOT NULL, -- 'checkbox', 'range', 'dropdown'
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Filter Attributes
CREATE TABLE filter_attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filter_id UUID NOT NULL REFERENCES filters(id) ON DELETE CASCADE,
    attribute_name VARCHAR(255) NOT NULL,
    attribute_value VARCHAR(255) NOT NULL,
    icon_url TEXT,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Filters (Many-to-Many relationship)
CREATE TABLE product_filters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    filter_attribute_id UUID NOT NULL REFERENCES filter_attributes(id) ON DELETE CASCADE,
    UNIQUE(product_id, filter_attribute_id)
);

-- Search History
CREATE TABLE search_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    search_query VARCHAR(255) NOT NULL,
    search_type VARCHAR(50), -- 'store', 'product', 'category'
    results_count INTEGER DEFAULT 0,
    location GEOGRAPHY(POINT, 4326),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- CART & WISHLIST
-- ============================================================

-- Shopping Cart
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity DECIMAL(10, 3) NOT NULL DEFAULT 1,
    attributes JSONB, -- Selected product attributes (DEPRECATED)
    special_instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Cart Item Variations (Selected variation in cart)
CREATE TABLE cart_item_variations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_item_id UUID NOT NULL REFERENCES cart_items(id) ON DELETE CASCADE,
    variation_id UUID NOT NULL REFERENCES product_variations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cart_item_id)
);

-- Cart Item Addons (Selected addons in cart)
CREATE TABLE cart_item_addons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cart_item_id UUID NOT NULL REFERENCES cart_items(id) ON DELETE CASCADE,
    addon_group_id UUID NOT NULL REFERENCES product_addon_groups(id) ON DELETE CASCADE,
    addon_product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Wishlist/Favorites
CREATE TABLE user_favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL, -- 'store', 'product'
    item_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, item_type, item_id)
);

-- ============================================================
-- REVIEWS & RATINGS
-- ============================================================

-- Product Reviews
CREATE TABLE product_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(255),
    comment TEXT,
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Store Reviews
CREATE TABLE store_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    is_verified_purchase BOOLEAN DEFAULT FALSE,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- PAYMENTS
-- ============================================================

-- Payment Methods
CREATE TABLE user_payment_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_type VARCHAR(50) NOT NULL, -- 'card', 'upi', 'netbanking', 'wallet'
    provider VARCHAR(100), -- 'razorpay', 'paytm', etc.
    token VARCHAR(255), -- Tokenized payment method
    last_four_digits VARCHAR(4),
    expiry_month INTEGER,
    expiry_year INTEGER,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Payment Transactions
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    payment_method_id UUID REFERENCES user_payment_methods(id),

    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    -- 'pending', 'processing', 'completed', 'failed', 'refunded'

    payment_gateway VARCHAR(100), -- 'razorpay', 'stripe', 'paytm'
    transaction_id VARCHAR(255) UNIQUE,
    gateway_response JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

-- User Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    -- 'order_update', 'promotion', 'new_product', 'delivery'

    related_entity_type VARCHAR(50), -- 'order', 'product', 'store'
    related_entity_id UUID,

    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Push Notification Tokens
CREATE TABLE push_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    token TEXT NOT NULL,
    platform VARCHAR(50) NOT NULL, -- 'ios', 'android', 'web'
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_id)
);

-- ============================================================
-- ANALYTICS & TRACKING
-- ============================================================

-- Product Views/Impressions
CREATE TABLE product_impressions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id UUID REFERENCES stores(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),

    impression_type VARCHAR(50), -- 'search', 'category', 'featured', 'related'
    search_query VARCHAR(255),
    position INTEGER, -- Position in list

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Store Impressions
CREATE TABLE store_impressions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),

    impression_type VARCHAR(50),
    position INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================

-- User indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Store indexes
CREATE INDEX idx_stores_location ON stores USING GIST(location);
CREATE INDEX idx_stores_city ON stores(city);
CREATE INDEX idx_stores_is_active ON stores(is_active);
CREATE INDEX idx_stores_rating ON stores(rating DESC);

-- Brand indexes
CREATE INDEX idx_brands_slug ON brands(slug);
CREATE INDEX idx_brands_normalized_name ON brands(normalized_name);
CREATE INDEX idx_brands_is_active ON brands(is_active);

-- Product indexes
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_brand_id ON products(brand_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_products_ean ON products(ean) WHERE ean IS NOT NULL;

-- Product matching indexes (for ERP integration)
CREATE INDEX idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_products_normalized_name_trgm ON products USING gin(normalized_name gin_trgm_ops);

-- Tax indexes
CREATE INDEX idx_taxes_tax_id ON taxes(tax_id);
CREATE INDEX idx_taxes_is_active ON taxes(is_active);

-- Store Product Tax indexes (for store-specific tax lookups)
CREATE INDEX idx_store_product_taxes_store_id ON store_product_taxes(store_id);
CREATE INDEX idx_store_product_taxes_store_product_id ON store_product_taxes(store_product_id);
CREATE INDEX idx_store_product_taxes_tax_id ON store_product_taxes(tax_id);
CREATE INDEX idx_store_product_taxes_is_active ON store_product_taxes(is_active);

-- Variation indexes
CREATE INDEX idx_variations_product_id ON product_variations(product_id);
CREATE INDEX idx_variations_is_active ON product_variations(is_active);

-- Store Product Mapping indexes (ERP Integration)
CREATE INDEX idx_store_product_mappings_store_id ON store_product_mappings(store_id);
CREATE INDEX idx_store_product_mappings_product_id ON store_product_mappings(product_id);
CREATE INDEX idx_store_product_mappings_external_id ON store_product_mappings(external_product_id);
CREATE INDEX idx_store_product_mappings_sync_source ON store_product_mappings(sync_source);
CREATE INDEX idx_store_product_mappings_is_active ON store_product_mappings(is_active);

-- Addon indexes
CREATE INDEX idx_addon_groups_product_id ON product_addon_groups(product_id);
CREATE INDEX idx_addon_group_items_group_id ON product_addon_group_items(addon_group_id);
CREATE INDEX idx_addon_group_items_store_product_id ON product_addon_group_items(addon_store_product_id);
CREATE INDEX idx_products_is_addon ON products(is_addon);

-- Order item variation/addon indexes
CREATE INDEX idx_order_item_variations_order_item_id ON order_item_variations(order_item_id);
CREATE INDEX idx_order_item_addons_order_item_id ON order_item_addons(order_item_id);

-- Cart item variation/addon indexes
CREATE INDEX idx_cart_item_variations_cart_item_id ON cart_item_variations(cart_item_id);
CREATE INDEX idx_cart_item_addons_cart_item_id ON cart_item_addons(cart_item_id);

-- Store products indexes
CREATE INDEX idx_store_products_store_id ON store_products(store_id);
CREATE INDEX idx_store_products_product_id ON store_products(product_id);
CREATE INDEX idx_store_products_external_id ON store_products(external_id) WHERE external_id IS NOT NULL;
CREATE INDEX idx_store_products_availability ON store_products(is_available, is_in_stock);

-- Order indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_store_id ON orders(store_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_order_number ON orders(order_number);

-- Order items indexes
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- Cart indexes
CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
CREATE INDEX idx_cart_items_store_id ON cart_items(store_id);

-- Search history indexes
CREATE INDEX idx_search_history_user_id ON search_history(user_id);
CREATE INDEX idx_search_history_query ON search_history(search_query);
CREATE INDEX idx_search_history_created_at ON search_history(created_at DESC);

-- Promotion indexes
CREATE INDEX idx_promotions_promo_code ON promotions(promo_code);
CREATE INDEX idx_promotions_dates ON promotions(start_date, end_date);
CREATE INDEX idx_promotions_is_active ON promotions(is_active);

-- Review indexes
CREATE INDEX idx_product_reviews_product_id ON product_reviews(product_id);
CREATE INDEX idx_product_reviews_user_id ON product_reviews(user_id);
CREATE INDEX idx_store_reviews_store_id ON store_reviews(store_id);

-- Address indexes
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_location ON user_addresses USING GIST(location);

-- Notification indexes
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- ============================================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_store_products_updated_at BEFORE UPDATE ON store_products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cart_items_updated_at BEFORE UPDATE ON cart_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_taxes_updated_at BEFORE UPDATE ON taxes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_variations_updated_at BEFORE UPDATE ON product_variations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_addon_groups_updated_at BEFORE UPDATE ON product_addon_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_taxes_updated_at BEFORE UPDATE ON taxes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================

-- Active stores with ratings
CREATE OR REPLACE VIEW v_active_stores AS
SELECT 
    s.*,
    COUNT(DISTINCT sr.id) as review_count,
    COALESCE(AVG(sr.rating), 0) as avg_rating
FROM stores s
LEFT JOIN store_reviews sr ON s.id = sr.store_id
WHERE s.is_active = TRUE
GROUP BY s.id;

-- Product catalog with store availability
CREATE OR REPLACE VIEW v_product_catalog AS
SELECT 
    p.*,
    c.name as category_name,
    c.slug as category_slug,
    COUNT(DISTINCT sp.store_id) as available_stores,
    MIN(sp.price) as min_price,
    MAX(sp.price) as max_price
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN store_products sp ON p.id = sp.product_id AND sp.is_available = TRUE
WHERE p.is_active = TRUE
GROUP BY p.id, c.name, c.slug;

-- User order summary
CREATE OR REPLACE VIEW v_user_order_summary AS
SELECT 
    u.id as user_id,
    u.email,
    u.first_name,
    u.last_name,
    COUNT(DISTINCT o.id) as total_orders,
    SUM(o.total_amount) as total_spent,
    MAX(o.created_at) as last_order_date
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id;

-- ============================================================
-- PRODUCT MATCHING ENGINE (ERP Integration)
-- ============================================================

-- Enable fuzzy matching extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Function to normalize product names for matching
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

-- Function to extract volume in milliliters
CREATE OR REPLACE FUNCTION extract_volume_ml(product_name TEXT)
RETURNS DECIMAL(10, 3) AS $$
DECLARE
    volume DECIMAL(10, 3);
    matches TEXT[];
BEGIN
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

-- Function to extract weight in grams
CREATE OR REPLACE FUNCTION extract_weight_g(product_name TEXT)
RETURNS DECIMAL(10, 3) AS $$
DECLARE
    weight DECIMAL(10, 3);
    matches TEXT[];
BEGIN
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

-- Function to find matching product using 3-layer strategy
-- Layer 1: Exact match (barcode, SKU, EAN, existing mapping)
-- Layer 2: Normalized match (name + volume/weight)
-- Layer 3: Fuzzy match (trigram similarity)
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
    
    -- Match by barcode
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
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-populate normalized fields
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

-- Trigger to auto-populate brand normalized_name
CREATE OR REPLACE FUNCTION update_brand_normalized_name()
RETURNS TRIGGER AS $$
BEGIN
    NEW.normalized_name := normalize_product_name(NEW.name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_brand_normalized_name
    BEFORE INSERT OR UPDATE OF name ON brands
    FOR EACH ROW
    EXECUTE FUNCTION update_brand_normalized_name();

-- Function to find or create brand with normalization
CREATE OR REPLACE FUNCTION find_or_create_brand(
    p_brand_name TEXT
)
RETURNS UUID AS $$
DECLARE
    v_brand_id UUID;
    v_normalized_name TEXT;
    v_slug TEXT;
BEGIN
    IF p_brand_name IS NULL OR TRIM(p_brand_name) = '' THEN
        RETURN NULL;
    END IF;
    
    v_normalized_name := normalize_product_name(p_brand_name);
    
    -- Try exact name match
    SELECT id INTO v_brand_id FROM brands WHERE name = p_brand_name LIMIT 1;
    IF v_brand_id IS NOT NULL THEN RETURN v_brand_id; END IF;
    
    -- Try normalized name match
    SELECT id INTO v_brand_id FROM brands WHERE normalized_name = v_normalized_name LIMIT 1;
    IF v_brand_id IS NOT NULL THEN RETURN v_brand_id; END IF;
    
    -- Create new brand
    v_slug := LOWER(REGEXP_REPLACE(p_brand_name, '[^a-zA-Z0-9]+', '-', 'g'));
    v_slug := TRIM(BOTH '-' FROM v_slug);
    IF EXISTS (SELECT 1 FROM brands WHERE slug = v_slug) THEN
        v_slug := v_slug || '-' || EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT;
    END IF;
    
    INSERT INTO brands (name, slug, normalized_name)
    VALUES (p_brand_name, v_slug, v_normalized_name)
    RETURNING id INTO v_brand_id;
    
    RETURN v_brand_id;
END;
$$ LANGUAGE plpgsql;

-- Add comments
COMMENT ON TABLE brands IS 'Normalized brand names for multi-ERP mapping. Different ERPs may use different brand names (Coca Cola, CocaCola, Coke) which map to the same brand.';
COMMENT ON FUNCTION find_or_create_brand IS 'Finds existing brand by name or normalized name, or creates new brand if not found. Handles brand name variations across ERPs.';
COMMENT ON FUNCTION find_matching_product IS 'Three-layer product matching for ERP integration: 1) Exact (barcode/SKU/EAN), 2) Normalized (name+size), 3) Fuzzy (similarity). Returns product_id, match_type, and confidence score.';
COMMENT ON FUNCTION normalize_product_name IS 'Normalizes product names by removing punctuation, filler words, and standardizing units for better matching.';
COMMENT ON FUNCTION extract_volume_ml IS 'Extracts volume in milliliters from product name (e.g., "1L" → 1000, "500ml" → 500).';
COMMENT ON FUNCTION extract_weight_g IS 'Extracts weight in grams from product name (e.g., "1kg" → 1000, "500g" → 500).';
COMMENT ON TABLE store_product_mappings IS 'Maps external ERP product identifiers to internal product IDs. Each store can have multiple ERP systems with different product IDs for the same product.';

-- ============================================================
-- END OF SCHEMA
-- ============================================================
