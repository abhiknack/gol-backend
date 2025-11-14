---
inclusion: always
---

# Grocery Superapp Database Schema

This project uses a comprehensive PostgreSQL database schema for a multi-store grocery platform.

## Schema Overview

The database schema is defined in `grocery_superapp_schema.sql` and includes:

### Core Entities
- **users** - Customer accounts with authentication
- **user_addresses** - Multiple delivery addresses per user with PostGIS location support
- **stores** - Grocery stores/shops with location, ratings, and business details
- **store_hours** - Operating hours for each store
- **categories** - Hierarchical product categories
- **products** - Product catalog with pricing, inventory, and media
- **product_images** - Multiple images per product
- **store_products** - Store-specific inventory and pricing
- **product_variations** - Product variations (sizes, flavors) with individual pricing
- **product_addon_groups** - Addon groups for products (toppings, sides, extras)
- **product_addon_group_items** - Links store_products to addon groups for store-specific addon pricing

### Orders & Transactions
- **orders** - Customer orders with delivery/pickup options
- **order_items** - Line items for each order
- **order_item_variations** - Snapshot of selected variations at order time
- **order_item_addons** - Snapshot of selected addons at order time
- **order_status_history** - Order status tracking

### Promotions & Discounts
- **promotions** - Discount codes and offers
- **promotion_stores** - Store-specific promotions
- **promotion_products** - Product-specific promotions
- **promotion_usage** - Promotion usage tracking

### Search & Filtering
- **filters** - Filter categories for product search
- **filter_attributes** - Filter options
- **product_filters** - Product-to-filter relationships
- **search_history** - User search tracking

### Cart & Wishlist
- **cart_items** - Shopping cart
- **cart_item_variations** - Selected variations in cart
- **cart_item_addons** - Selected addons in cart
- **user_favorites** - Wishlist/favorites

### Reviews & Ratings
- **product_reviews** - Product reviews and ratings
- **store_reviews** - Store reviews and ratings

### Payments
- **user_payment_methods** - Saved payment methods
- **payments** - Payment transactions

### Notifications
- **notifications** - User notifications
- **push_tokens** - Push notification device tokens

### Analytics
- **product_impressions** - Product view tracking
- **store_impressions** - Store view tracking

## Key Features

### Product Variations & Addons
- **Variations**: Products can have multiple variations (e.g., Small, Medium, Large)
  - Each variation has its own price and inventory
  - One variation can be selected per product in cart/order
- **Addons**: Products can have addon groups (e.g., "Choose Your Side", "Extra Toppings")
  - Addon groups contain other products (marked with `is_addon=true`) as addon options
  - Products can be both regular products AND addons (dual purpose)
  - Supports min/max selections, required/optional groups
  - Price override option for addons

### PostGIS Integration
- Location-based queries using `GEOGRAPHY(POINT, 4326)`
- Spatial indexes for efficient location searches
- Used in: `stores`, `user_addresses`, `search_history`

### UUID Primary Keys
- All tables use UUID for primary keys
- Generated using `uuid_generate_v4()`

### Timestamps
- `created_at` and `updated_at` on most tables
- Automatic `updated_at` triggers

### Soft Deletes
- Most tables use `is_active` flags instead of hard deletes
- Maintains referential integrity

### Order Snapshots
- Variations and addons are snapshotted at order time
- Preserves order details even if products/variations/addons are deleted

## Important Indexes

Performance-critical indexes include:
- Location-based: `idx_stores_location`, `idx_user_addresses_location`
- Search: `idx_products_slug`, `idx_stores_city`
- Orders: `idx_orders_user_id`, `idx_orders_status`
- Cart: `idx_cart_items_user_id`
- Variations: `idx_variations_product_id`, `idx_variations_is_active`
- Addons: `idx_addon_groups_product_id`, `idx_addon_group_items_group_id`
- Cart variations/addons: `idx_cart_item_variations_cart_item_id`, `idx_cart_item_addons_cart_item_id`
- Order variations/addons: `idx_order_item_variations_order_item_id`, `idx_order_item_addons_order_item_id`

## Views

Pre-defined views for common queries:
- `v_active_stores` - Active stores with ratings
- `v_product_catalog` - Products with store availability
- `v_user_order_summary` - User order statistics

## Database Connection

Current setup uses:
- **Database**: `middleware_db`
- **Connection**: PostgreSQL with pgx driver
- **URL**: `postgresql://postgres:postgres@postgres:5432/middleware_db?sslmode=disable`

## Schema File Location

The complete schema is in: `grocery_superapp_schema.sql`

## Usage Notes

When implementing features:
1. Use UUID for all entity references
2. Include location data where applicable (PostGIS)
3. Maintain `is_active` flags for soft deletes
4. Use appropriate indexes for queries
5. Follow the established naming conventions
6. Use JSONB for flexible attributes
7. Include proper foreign key constraints

### Working with Variations & Addons
1. **Variations**: Create variations directly on products
   - Each variation has complete pricing (not price modifiers)
   - Use `is_default` to mark default selection
   - Track inventory per variation
2. **Addons**: Create addon groups on products, then link store_products as addon options
   - Mark products with `is_addon=true` to indicate they can be used as addons
   - Use `product_addon_group_items` to link store_products (not products) to addon groups
   - This allows store-specific addon pricing and availability
   - Optional `price_override` to override store_product's price
   - Set `min_selections` and `max_selections` for validation
   - Products can be both regular products AND addons (dual purpose)
3. **Cart/Orders**: Link variations and addons to cart/order items
   - One variation per cart/order item
   - Multiple addons per cart/order item
   - Orders snapshot variation/addon details for historical accuracy

## Migration Strategy

For schema changes:
1. Create migration files in `migrations/` directory
2. Use `golang-migrate` for version control
3. Test migrations on development database first
4. Include both `up` and `down` migrations

## Related Files

- Schema: `grocery_superapp_schema.sql`
- Repository: `internal/repository/postgres.go`
- Config: `config/config.go` (DatabaseConfig)
- Init Script: `init-postgres.sql` (sample data)
