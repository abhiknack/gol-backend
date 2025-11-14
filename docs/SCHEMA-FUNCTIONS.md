# Database Functions in Schema

## Overview

The `grocery_superapp_schema.sql` file includes all necessary database functions for the application.

## Functions List

### 1. `update_updated_at_column()`
**Purpose:** Trigger function to automatically update `updated_at` timestamp  
**Returns:** TRIGGER  
**Usage:** Attached to tables that need automatic timestamp updates

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

---

### 2. `normalize_product_name(product_name TEXT)`
**Purpose:** Normalizes product names for better matching  
**Returns:** TEXT  
**Usage:** Used by product matching engine

**Normalization 