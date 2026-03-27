-- Migration to add discount restrictions
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS restriction_type character varying(50) DEFAULT 'ALL' NOT NULL;
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS target_ids uuid[];

-- restriction_type can be:
-- 'ALL' - Works for everything
-- 'ITEM_CATEGORY' - Works for specific product categories
-- 'GAME_CATEGORY' - Works for specific game categories (this might be same as above depending on implementation)
-- 'GAME' - Works for a specific game
-- 'PRODUCT' - Works for specific products
