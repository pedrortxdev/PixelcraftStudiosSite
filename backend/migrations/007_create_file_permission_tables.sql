-- Migration 007: Create file permission association tables
-- Solves the "JSON Lazy Schema" problem by creating proper relational tables
-- for file permissions instead of storing JSON blobs

-- Create table for file-role permissions
CREATE TABLE IF NOT EXISTS file_permission_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    role_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(file_id, role_name)
);

-- Create table for file-product permissions
CREATE TABLE IF NOT EXISTS file_permission_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(file_id, product_id)
);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_file_permission_roles_file_id ON file_permission_roles(file_id);
CREATE INDEX IF NOT EXISTS idx_file_permission_roles_role_name ON file_permission_roles(role_name);
CREATE INDEX IF NOT EXISTS idx_file_permission_products_file_id ON file_permission_products(file_id);
CREATE INDEX IF NOT EXISTS idx_file_permission_products_product_id ON file_permission_products(product_id);

-- Add comment to document purpose
COMMENT ON TABLE file_permission_roles IS 'Association table for file access permissions by user role (replaces JSON allowed_roles)';
COMMENT ON TABLE file_permission_products IS 'Association table for file access permissions by product ownership (replaces JSON allowed_product_ids)';

-- Migration to populate new tables from existing JSON data (optional - can be run separately)
-- This migrates existing file permissions from JSON columns to relational tables
DO $$
BEGIN
    -- Migrate allowed_roles from JSON to file_permission_roles
    INSERT INTO file_permission_roles (file_id, role_name)
    SELECT 
        f.id,
        jsonb_array_elements_text(f.allowed_roles::jsonb) as role_name
    FROM files f
    WHERE f.allowed_roles IS NOT NULL 
      AND f.allowed_roles != 'null'
      AND jsonb_array_length(f.allowed_roles::jsonb) > 0
    ON CONFLICT (file_id, role_name) DO NOTHING;

    -- Migrate allowed_product_ids from JSON to file_permission_products
    INSERT INTO file_permission_products (file_id, product_id)
    SELECT 
        f.id,
        jsonb_array_elements_text(f.allowed_product_ids::jsonb)::uuid as product_id
    FROM files f
    WHERE f.allowed_product_ids IS NOT NULL 
      AND f.allowed_product_ids != 'null'
      AND jsonb_array_length(f.allowed_product_ids::jsonb) > 0
    ON CONFLICT (file_id, product_id) DO NOTHING;
END $$;
