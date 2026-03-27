-- Default Role Permissions Configuration
-- This migration sets up default permissions for each role

-- Clear existing permissions (optional - comment out if you want to keep custom permissions)
-- TRUNCATE TABLE role_permissions;

-- SUPPORT Role: Limited access to support tickets and own emails
INSERT INTO role_permissions (role, resource, action) VALUES
('SUPPORT', 'SUPPORT', 'VIEW'),
('SUPPORT', 'SUPPORT', 'CREATE'),
('SUPPORT', 'SUPPORT', 'EDIT'),
('SUPPORT', 'EMAILS', 'VIEW'),
('SUPPORT', 'EMAILS', 'CREATE'),
('SUPPORT', 'DASHBOARD', 'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ADMIN Role: View-only access to most resources
INSERT INTO role_permissions (role, resource, action) VALUES
('ADMIN', 'USERS', 'VIEW'),
('ADMIN', 'PRODUCTS', 'VIEW'),
('ADMIN', 'ORDERS', 'VIEW'),
('ADMIN', 'TRANSACTIONS', 'VIEW'),
('ADMIN', 'SUPPORT', 'VIEW'),
('ADMIN', 'EMAILS', 'VIEW'),
('ADMIN', 'FILES', 'VIEW'),
('ADMIN', 'GAMES', 'VIEW'),
('ADMIN', 'CATEGORIES', 'VIEW'),
('ADMIN', 'PLANS', 'VIEW'),
('ADMIN', 'DASHBOARD', 'VIEW'),
('ADMIN', 'SETTINGS', 'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;

-- DEVELOPMENT Role: Can edit products, plans, games, and categories
INSERT INTO role_permissions (role, resource, action) VALUES
-- View permissions
('DEVELOPMENT', 'USERS', 'VIEW'),
('DEVELOPMENT', 'PRODUCTS', 'VIEW'),
('DEVELOPMENT', 'ORDERS', 'VIEW'),
('DEVELOPMENT', 'TRANSACTIONS', 'VIEW'),
('DEVELOPMENT', 'SUPPORT', 'VIEW'),
('DEVELOPMENT', 'EMAILS', 'VIEW'),
('DEVELOPMENT', 'FILES', 'VIEW'),
('DEVELOPMENT', 'GAMES', 'VIEW'),
('DEVELOPMENT', 'CATEGORIES', 'VIEW'),
('DEVELOPMENT', 'PLANS', 'VIEW'),
('DEVELOPMENT', 'DASHBOARD', 'VIEW'),
('DEVELOPMENT', 'SETTINGS', 'VIEW'),
-- Edit permissions for catalog
('DEVELOPMENT', 'PRODUCTS', 'CREATE'),
('DEVELOPMENT', 'PRODUCTS', 'EDIT'),
('DEVELOPMENT', 'PRODUCTS', 'DELETE'),
('DEVELOPMENT', 'PLANS', 'CREATE'),
('DEVELOPMENT', 'PLANS', 'EDIT'),
('DEVELOPMENT', 'PLANS', 'DELETE'),
('DEVELOPMENT', 'GAMES', 'CREATE'),
('DEVELOPMENT', 'GAMES', 'EDIT'),
('DEVELOPMENT', 'GAMES', 'DELETE'),
('DEVELOPMENT', 'CATEGORIES', 'CREATE'),
('DEVELOPMENT', 'CATEGORIES', 'EDIT'),
('DEVELOPMENT', 'CATEGORIES', 'DELETE'),
('DEVELOPMENT', 'FILES', 'CREATE'),
('DEVELOPMENT', 'FILES', 'EDIT'),
('DEVELOPMENT', 'FILES', 'DELETE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ENGINEERING Role: Full access except roles management
INSERT INTO role_permissions (role, resource, action) VALUES
-- Manage permissions (all actions)
('ENGINEERING', 'USERS', 'MANAGE'),
('ENGINEERING', 'PRODUCTS', 'MANAGE'),
('ENGINEERING', 'ORDERS', 'MANAGE'),
('ENGINEERING', 'TRANSACTIONS', 'MANAGE'),
('ENGINEERING', 'SUPPORT', 'MANAGE'),
('ENGINEERING', 'EMAILS', 'MANAGE'),
('ENGINEERING', 'FILES', 'MANAGE'),
('ENGINEERING', 'GAMES', 'MANAGE'),
('ENGINEERING', 'CATEGORIES', 'MANAGE'),
('ENGINEERING', 'PLANS', 'MANAGE'),
('ENGINEERING', 'DASHBOARD', 'MANAGE'),
('ENGINEERING', 'SETTINGS', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- DIRECTION Role: Full access to everything including roles
INSERT INTO role_permissions (role, resource, action) VALUES
-- Manage permissions (all actions)
('DIRECTION', 'USERS', 'MANAGE'),
('DIRECTION', 'ROLES', 'MANAGE'),
('DIRECTION', 'PRODUCTS', 'MANAGE'),
('DIRECTION', 'ORDERS', 'MANAGE'),
('DIRECTION', 'TRANSACTIONS', 'MANAGE'),
('DIRECTION', 'SUPPORT', 'MANAGE'),
('DIRECTION', 'EMAILS', 'MANAGE'),
('DIRECTION', 'FILES', 'MANAGE'),
('DIRECTION', 'GAMES', 'MANAGE'),
('DIRECTION', 'CATEGORIES', 'MANAGE'),
('DIRECTION', 'PLANS', 'MANAGE'),
('DIRECTION', 'DASHBOARD', 'MANAGE'),
('DIRECTION', 'SETTINGS', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- Log the migration
SELECT 'Default role permissions configured successfully' AS status;
