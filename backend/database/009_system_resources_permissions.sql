-- System Resources Permission Configuration
-- This migration adds SYSTEM resource permissions for viewing server metrics

-- Add SYSTEM resource permissions for each role
-- SUPPORT: No access to system metrics
-- ADMIN: View-only access to system metrics
-- DEVELOPMENT+: Full access to system metrics

-- ADMIN Role: View-only access to system metrics
INSERT INTO role_permissions (role, resource, action) VALUES
('ADMIN', 'SYSTEM', 'VIEW')
ON CONFLICT (role, resource, action) DO NOTHING;

-- DEVELOPMENT Role: Full access to system metrics
INSERT INTO role_permissions (role, resource, action) VALUES
('DEVELOPMENT', 'SYSTEM', 'VIEW'),
('DEVELOPMENT', 'SYSTEM', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- ENGINEERING Role: Full access to system metrics
INSERT INTO role_permissions (role, resource, action) VALUES
('ENGINEERING', 'SYSTEM', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- DIRECTION Role: Full access to system metrics
INSERT INTO role_permissions (role, resource, action) VALUES
('DIRECTION', 'SYSTEM', 'MANAGE')
ON CONFLICT (role, resource, action) DO NOTHING;

-- Log the migration
SELECT 'System resource permissions configured successfully' AS status;
