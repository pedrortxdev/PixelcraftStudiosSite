-- Permission System Enhancements
-- Adds: Permission inheritance, audit log, custom roles, and notifications

-- 1. Add permission inheritance flag to roles
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS is_inherited BOOLEAN DEFAULT FALSE;
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS inherited_from VARCHAR(50);

-- 2. Create permission audit log table
CREATE TABLE IF NOT EXISTS permission_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    operation VARCHAR(20) NOT NULL, -- 'ADD', 'REMOVE', 'INHERIT'
    performed_by UUID,
    performed_at TIMESTAMP DEFAULT NOW(),
    old_value JSONB,
    new_value JSONB,
    reason TEXT,
    FOREIGN KEY (performed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_permission_audit_role ON permission_audit_log(role);
CREATE INDEX IF NOT EXISTS idx_permission_audit_performed_at ON permission_audit_log(performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_permission_audit_performed_by ON permission_audit_log(performed_by);

-- 3. Create custom roles table
CREATE TABLE IF NOT EXISTS custom_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#999999',
    hierarchy_level INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CHECK (hierarchy_level >= 1 AND hierarchy_level <= 10)
);

CREATE INDEX IF NOT EXISTS idx_custom_roles_active ON custom_roles(is_active);
CREATE INDEX IF NOT EXISTS idx_custom_roles_hierarchy ON custom_roles(hierarchy_level);

-- 4. Create permission templates table (for export/import)
CREATE TABLE IF NOT EXISTS permission_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(100) NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL,
    created_by UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    is_public BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_permission_templates_public ON permission_templates(is_public);

-- 5. Create permission notifications table
CREATE TABLE IF NOT EXISTS permission_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    notification_type VARCHAR(50) NOT NULL, -- 'PERMISSION_ADDED', 'PERMISSION_REMOVED', 'ROLE_ASSIGNED', 'ROLE_REMOVED'
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_permission_notifications_user ON permission_notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_permission_notifications_created ON permission_notifications(created_at DESC);

-- 6. Function to inherit permissions from lower roles
CREATE OR REPLACE FUNCTION inherit_permissions_from_role(
    target_role VARCHAR(50),
    source_role VARCHAR(50)
) RETURNS INT AS $$
DECLARE
    inserted_count INT := 0;
BEGIN
    -- Insert permissions from source role to target role (if not exists)
    INSERT INTO role_permissions (role, resource, action, is_inherited, inherited_from)
    SELECT 
        target_role,
        resource,
        action,
        TRUE,
        source_role
    FROM role_permissions
    WHERE role = source_role
    ON CONFLICT (role, resource, action) DO NOTHING;
    
    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    RETURN inserted_count;
END;
$$ LANGUAGE plpgsql;

-- 7. Function to remove inherited permissions
CREATE OR REPLACE FUNCTION remove_inherited_permissions(
    target_role VARCHAR(50)
) RETURNS INT AS $$
DECLARE
    deleted_count INT := 0;
BEGIN
    DELETE FROM role_permissions
    WHERE role = target_role AND is_inherited = TRUE;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 8. Function to log permission changes
CREATE OR REPLACE FUNCTION log_permission_change() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO permission_audit_log (role, resource, action, operation, new_value)
        VALUES (NEW.role, NEW.resource, NEW.action, 'ADD', row_to_json(NEW)::jsonb);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO permission_audit_log (role, resource, action, operation, old_value)
        VALUES (OLD.role, OLD.resource, OLD.action, 'REMOVE', row_to_json(OLD)::jsonb);
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO permission_audit_log (role, resource, action, operation, old_value, new_value)
        VALUES (NEW.role, NEW.resource, NEW.action, 'UPDATE', row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 9. Create trigger for permission audit
DROP TRIGGER IF EXISTS permission_audit_trigger ON role_permissions;
CREATE TRIGGER permission_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON role_permissions
FOR EACH ROW EXECUTE FUNCTION log_permission_change();

-- 10. Setup permission inheritance for default roles
-- DIRECTION inherits from ENGINEERING
SELECT inherit_permissions_from_role('DIRECTION', 'ENGINEERING');

-- ENGINEERING inherits from DEVELOPMENT
SELECT inherit_permissions_from_role('ENGINEERING', 'DEVELOPMENT');

-- DEVELOPMENT inherits from ADMIN
SELECT inherit_permissions_from_role('DEVELOPMENT', 'ADMIN');

-- ADMIN inherits from SUPPORT (VIEW permissions)
SELECT inherit_permissions_from_role('ADMIN', 'SUPPORT');

-- Log the migration
SELECT 'Permission system enhancements applied successfully' AS status;
