# Addon Architecture

## Overview

Product addons allow customers to customize their orders by adding extra items (toppings, sides, extras). The addon system uses `store_products` references to support store-specific pricing and availability.

## Why Store Products?

### Problem with Product-level Addons

If addons referenced `products` directly:
- ❌ All stores would have same addon prices
- ❌ Can't handle store-specific addon availability
- ❌ Can't track addon stock levels per store
- ❌ Can't support regional addon variations

### Solution: Store Product Addons

By referencing `store_products`:
- ✅ Each store can have different addon prices
- ✅ Store-specific addon availability
- ✅ Track addon stock levels per store
- ✅ Support regional variations (e.g., different toppings by location)

## Database Schema

### Addon Groups

Defines what addon options are available for a product:

```sql
CREATE TABLE product_addon_groups (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id),  -- Base product
    name VARCHAR(100),                         -- 'Extra Toppings'
    display_name VARCHAR(255),                 -- 'Add Extra Toppings'
    is_required BOOLEAN DEFAULT FALSE,
    min_selections INTEGER DEFAULT 0,
    max_selections INTEGER,                    -- NULL = unlimited
    display_order INTEGER DEFAULT 0
);
```

### Addon Items

Links store products as addon options:

```sql
CREATE TABLE product_addon_group_items (
    id UUID PRIMARY KEY,
    addon_group_id UUID REFERENCES product_addon_groups(id),
    addon_store_product_id UUID REFERENCES store_products(id),  -- Store-specific!
    price_override DECIMAL(10, 2),                              -- Optional override
    use_sale_price BOOLEAN DEFAULT TRUE,
    display_order INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT FALSE,
    UNIQUE(addon_group_id, addon_store_product_id)
);
```

## Architecture Flow

```
Product (Pizza)
    ↓
Addon Group (Extra Toppings)
    ↓
Addon Items → Store Products (Cheese, Pepperoni, Mushrooms)
    ↓
Store-specific pricing & availability
```

## Usage Examples

### Example 1: Pizza with Toppings

**Setup:**

```sql
-- 1. Create base product (Pizza)
INSERT INTO products (id, sku, name, slug, base_price, is_customizable)
VALUES ('pizza-uuid', 'PIZZA-001', 'Margherita Pizza', 'margherita-pizza', 299.00, true);

-- 2. Create addon products (Toppings)
INSERT INTO products (id, sku, name, slug, base_price, is_addon)
VALUES 
    ('cheese-uuid', 'TOP-CHEESE', 'Extra Cheese', 'extra-cheese', 50.00, true),
    ('pepperoni-uuid', 'TOP-PEPPERONI', 'Pepperoni', 'pepperoni', 75.00, true),
    ('mushroom-uuid', 'TOP-MUSHROOM', 'Mushrooms', 'mushrooms', 40.00, true);

-- 3. Create store products for each store
-- Store A (Mumbai)
INSERT INTO store_products (store_id, product_id, price)
VALUES 
    ('mumbai-store-uuid', 'pizza-uuid', 299.00),
    ('mumbai-store-uuid', 'cheese-uuid', 50.00),
    ('mumbai-store-uuid', 'pepperoni-uuid', 75.00),
    ('mumbai-store-uuid', 'mushroom-uuid', 40.00);

-- Store B (Delhi) - Different prices!
INSERT INTO store_products (store_id, product_id, price)
VALUES 
    ('delhi-store-uuid', 'pizza-uuid', 279.00),
    ('delhi-store-uuid', 'cheese-uuid', 45.00),
    ('delhi-store-uuid', 'pepperoni-uuid', 70.00),
    ('delhi-store-uuid', 'mushroom-uuid', 35.00);

-- 4. Create addon group
INSERT INTO product_addon_groups (id, product_id, name, display_name, min_selections, max_selections)
VALUES ('toppings-group-uuid', 'pizza-uuid', 'Extra Toppings', 'Add Extra Toppings', 0, 5);

-- 5. Link store products as addon options (Mumbai)
INSERT INTO product_addon_group_items (addon_group_id, addon_store_product_id, display_order)
SELECT 
    'toppings-group-uuid',
    sp.id,
    ROW_NUMBER() OVER (ORDER BY p.name)
FROM store_products sp
JOIN products p ON sp.product_id = p.id
WHERE sp.store_id = 'mumbai-store-uuid'
  AND p.is_addon = true;

-- 6. Link store products as addon options (Delhi)
INSERT INTO product_addon_group_items (addon_group_id, addon_store_product_id, display_order)
SELECT 
    'toppings-group-uuid',
    sp.id,
    ROW_NUMBER() OVER (ORDER BY p.name)
FROM store_products sp
JOIN products p ON sp.product_id = p.id
WHERE sp.store_id = 'delhi-store-uuid'
  AND p.is_addon = true;
```

**Result:**
- Mumbai store: Cheese ₹50, Pepperoni ₹75, Mushrooms ₹40
- Delhi store: Cheese ₹45, Pepperoni ₹70, Mushrooms ₹35

### Example 2: Burger with Sides

```sql
-- 1. Create burger and sides
INSERT INTO products (id, sku, name, slug, base_price, is_customizable, is_addon)
VALUES 
    ('burger-uuid', 'BURGER-001', 'Classic Burger', 'classic-burger', 199.00, true, false),
    ('fries-uuid', 'SIDE-FRIES', 'French Fries', 'french-fries', 60.00, false, true),
    ('coleslaw-uuid', 'SIDE-COLESLAW', 'Coleslaw', 'coleslaw', 50.00, false, true),
    ('onion-rings-uuid', 'SIDE-ONION', 'Onion Rings', 'onion-rings', 70.00, false, true);

-- 2. Create addon group
INSERT INTO product_addon_groups (id, product_id, name, display_name, is_required, min_selections, max_selections)
VALUES ('sides-group-uuid', 'burger-uuid', 'Choose Your Side', 'Choose Your Side', true, 1, 1);

-- 3. Link store products as addon options
INSERT INTO product_addon_group_items (addon_group_id, addon_store_product_id, display_order)
SELECT 
    'sides-group-uuid',
    sp.id,
    ROW_NUMBER() OVER (ORDER BY p.name)
FROM store_products sp
JOIN products p ON sp.product_id = p.id
WHERE sp.store_id = 'store-uuid'
  AND p.id IN ('fries-uuid', 'coleslaw-uuid', 'onion-rings-uuid');
```

### Example 3: Price Override

```sql
-- Combo deal: Fries normally ₹60, but only ₹30 when added to burger
INSERT INTO product_addon_group_items (
    addon_group_id, 
    addon_store_product_id, 
    price_override
)
SELECT 
    'sides-group-uuid',
    sp.id,
    30.00  -- Override price
FROM store_products sp
WHERE sp.store_id = 'store-uuid'
  AND sp.product_id = 'fries-uuid';
```

## Querying Addons

### Get Available Addons for Product in Store

```sql
SELECT 
    pag.name as group_name,
    pag.display_name,
    pag.is_required,
    pag.min_selections,
    pag.max_selections,
    p.name as addon_name,
    COALESCE(pagi.price_override, sp.price) as addon_price,
    sp.is_in_stock,
    sp.is_available
FROM product_addon_groups pag
JOIN product_addon_group_items pagi ON pag.id = pagi.addon_group_id
JOIN store_products sp ON pagi.addon_store_product_id = sp.id
JOIN products p ON sp.product_id = p.id
WHERE pag.product_id = 'pizza-uuid'
  AND sp.store_id = 'mumbai-store-uuid'
  AND pag.is_active = true
  AND pagi.is_active = true
  AND sp.is_available = true
ORDER BY pag.display_order, pagi.display_order;
```

### Calculate Order Total with Addons

```sql
-- Base product + selected addons
SELECT 
    sp_base.price as base_price,
    SUM(COALESCE(pagi.price_override, sp_addon.price)) as addons_total,
    sp_base.price + SUM(COALESCE(pagi.price_override, sp_addon.price)) as total_price
FROM store_products sp_base
LEFT JOIN product_addon_group_items pagi ON pagi.addon_store_product_id IN (
    -- Selected addon store_product_ids
    'addon-sp-1-uuid', 'addon-sp-2-uuid'
)
LEFT JOIN store_products sp_addon ON pagi.addon_store_product_id = sp_addon.id
WHERE sp_base.id = 'base-store-product-uuid'
GROUP BY sp_base.id, sp_base.price;
```

## Cart & Order Integration

### Adding to Cart with Addons

```sql
-- 1. Add base item to cart
INSERT INTO cart_items (user_id, store_product_id, quantity)
VALUES ('user-uuid', 'pizza-store-product-uuid', 1)
RETURNING id;

-- 2. Add selected addons
INSERT INTO cart_item_addons (cart_item_id, addon_store_product_id, quantity, price_at_time)
SELECT 
    'cart-item-uuid',
    pagi.addon_store_product_id,
    1,
    COALESCE(pagi.price_override, sp.price)
FROM product_addon_group_items pagi
JOIN store_products sp ON pagi.addon_store_product_id = sp.id
WHERE pagi.addon_store_product_id IN (
    'cheese-sp-uuid', 'pepperoni-sp-uuid'  -- Selected addons
);
```

### Order Snapshot

When order is placed, addon details are snapshotted:

```sql
INSERT INTO order_item_addons (
    order_item_id,
    addon_store_product_id,
    addon_name,
    addon_price,
    quantity
)
SELECT 
    'order-item-uuid',
    cia.addon_store_product_id,
    p.name,
    cia.price_at_time,
    cia.quantity
FROM cart_item_addons cia
JOIN store_products sp ON cia.addon_store_product_id = sp.id
JOIN products p ON sp.product_id = p.id
WHERE cia.cart_item_id = 'cart-item-uuid';
```

## Benefits

### 1. Store-Specific Pricing

```
Mumbai Store:
- Pizza: ₹299
- Extra Cheese: ₹50
- Total: ₹349

Delhi Store:
- Pizza: ₹279
- Extra Cheese: ₹45
- Total: ₹324
```

### 2. Regional Variations

```
North India Stores:
- Paneer Topping: ₹60

South India Stores:
- Paneer Topping: ₹50
- Coconut Topping: ₹40 (not available in North)
```

### 3. Stock Management

```sql
-- Check addon availability
SELECT 
    p.name,
    sp.stock_quantity,
    sp.is_in_stock
FROM product_addon_group_items pagi
JOIN store_products sp ON pagi.addon_store_product_id = sp.id
JOIN products p ON sp.product_id = p.id
WHERE pagi.addon_group_id = 'toppings-group-uuid'
  AND sp.store_id = 'store-uuid';
```

### 4. Dynamic Pricing

```sql
-- Happy hour: 50% off all toppings
UPDATE product_addon_group_items pagi
SET price_override = sp.price * 0.5
FROM store_products sp
WHERE pagi.addon_store_product_id = sp.id
  AND sp.store_id = 'store-uuid'
  AND EXISTS (
      SELECT 1 FROM product_addon_groups pag
      WHERE pag.id = pagi.addon_group_id
        AND pag.name = 'Extra Toppings'
  );
```

## Migration from Product-based Addons

If you have existing addon data referencing products:

```sql
-- Create store_products for addon products
INSERT INTO store_products (store_id, product_id, price, is_available)
SELECT 
    s.id as store_id,
    p.id as product_id,
    p.base_price,
    true
FROM products p
CROSS JOIN stores s
WHERE p.is_addon = true
ON CONFLICT (store_id, product_id) DO NOTHING;

-- Update addon items to reference store_products
UPDATE product_addon_group_items pagi
SET addon_store_product_id = sp.id
FROM store_products sp
WHERE sp.product_id = pagi.addon_product_id  -- old column
  AND sp.store_id = (
      -- Determine store context from addon group's product
      SELECT sp2.store_id 
      FROM product_addon_groups pag
      JOIN store_products sp2 ON pag.product_id = sp2.product_id
      WHERE pag.id = pagi.addon_group_id
      LIMIT 1
  );
```

## Best Practices

1. **Always use store_products** - Never reference products directly in addon items
2. **Create store_products first** - Before linking as addons
3. **Use price_override sparingly** - Only for special deals/combos
4. **Check availability** - Query `is_available` and `is_in_stock` before showing addons
5. **Snapshot prices** - Store addon prices in cart/order for historical accuracy
6. **Validate selections** - Respect `min_selections` and `max_selections` constraints

This architecture provides maximum flexibility for multi-store operations while maintaining data integrity and supporting complex pricing scenarios!
