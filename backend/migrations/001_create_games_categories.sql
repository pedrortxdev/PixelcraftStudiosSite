-- Migration: Create games and categories tables
-- Run this migration to add game-based filtering support

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    icon_url TEXT,
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Categories table (subcategories within games)
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID REFERENCES games(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    display_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(game_id, slug)
);

-- Add game_id and category_id to products table
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS game_id UUID REFERENCES games(id),
ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES categories(id),
ADD COLUMN IF NOT EXISTS file_id UUID REFERENCES files(id),
ADD COLUMN IF NOT EXISTS download_url_encrypted BYTEA;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_products_game_id ON products(game_id);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_categories_game_id ON categories(game_id);

-- Seed initial games
INSERT INTO games (name, slug, icon_url, display_order) VALUES
    ('Minecraft', 'minecraft', NULL, 1),
    ('Roblox', 'roblox', NULL, 2),
    ('FiveM', 'fivem', NULL, 3),
    ('Outros', 'outros', NULL, 99)
ON CONFLICT (slug) DO NOTHING;

-- Seed initial categories for Minecraft
INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Plugins', 'plugins', 1 FROM games g WHERE g.slug = 'minecraft'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Mods', 'mods', 2 FROM games g WHERE g.slug = 'minecraft'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Mapas', 'mapas', 3 FROM games g WHERE g.slug = 'minecraft'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Texture Packs', 'texture-packs', 4 FROM games g WHERE g.slug = 'minecraft'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Servidores Prontos', 'servidores-prontos', 5 FROM games g WHERE g.slug = 'minecraft'
ON CONFLICT (game_id, slug) DO NOTHING;

-- Seed categories for FiveM
INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Scripts', 'scripts', 1 FROM games g WHERE g.slug = 'fivem'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Veículos', 'veiculos', 2 FROM games g WHERE g.slug = 'fivem'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Mapas', 'mapas', 3 FROM games g WHERE g.slug = 'fivem'
ON CONFLICT (game_id, slug) DO NOTHING;

-- Seed categories for Roblox
INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Scripts', 'scripts', 1 FROM games g WHERE g.slug = 'roblox'
ON CONFLICT (game_id, slug) DO NOTHING;

INSERT INTO categories (game_id, name, slug, display_order) 
SELECT g.id, 'Modelos', 'modelos', 2 FROM games g WHERE g.slug = 'roblox'
ON CONFLICT (game_id, slug) DO NOTHING;
