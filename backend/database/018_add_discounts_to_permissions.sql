-- Migration 018: Add Discounts Resource to Permissions System
DO $$ BEGIN
    ALTER TYPE resource_type ADD VALUE IF NOT EXISTS 'DISCOUNTS';
EXCEPTION WHEN others THEN
    NULL;
END $$;

COMMIT;
BEGIN;

-- Adiciona MANAGE para os cargos que cuidam de catálogo/produtos
INSERT INTO role_permissions (role, resource, action) VALUES
    ('DEVELOPMENT', 'DISCOUNTS', 'MANAGE'),
    ('ENGINEERING', 'DISCOUNTS', 'MANAGE'),
    ('DIRECTION',   'DISCOUNTS', 'MANAGE'),
    ('SUPPORT',     'DISCOUNTS', 'VIEW')
ON CONFLICT DO NOTHING;
