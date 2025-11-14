-- Add unique constraints for ON CONFLICT support in upsert operations
-- This migration adds constraints to product_images and product_variations tables

-- 1. Product Images: Remove duplicates and add unique constraint
DELETE FROM product_images a
USING product_images b
WHERE a.id > b.id
  AND a.product_id = b.product_id
  AND a.image_url = b.image_url;

ALTER TABLE product_images
ADD CONSTRAINT product_images_product_id_image_url_key
UNIQUE (product_id, image_url);

CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);

-- 2. Product Variations: Remove duplicates and add unique constraint
DELETE FROM product_variations a
USING product_variations b
WHERE a.id > b.id
  AND a.product_id = b.product_id
  AND a.name = b.name;

ALTER TABLE product_variations
ADD CONSTRAINT product_variations_product_id_name_key
UNIQUE (product_id, name);

CREATE INDEX IF NOT EXISTS idx_product_variations_product_id ON product_variations(product_id);
