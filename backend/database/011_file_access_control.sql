-- Migration: Pixelcraft File Access Control System
-- Description: Adds comprehensive file access control with role-based and product-based permissions
-- Date: 2026-02-28

-- 1. Add access control columns to files table
ALTER TABLE files ADD COLUMN IF NOT EXISTS access_type VARCHAR(20) DEFAULT 'PRIVATE'::VARCHAR(20);
-- access_type: 'PUBLIC' (anyone with link), 'PRIVATE' (only buyers/roles), 'ROLE' (specific roles only)

ALTER TABLE files ADD COLUMN IF NOT EXISTS required_role VARCHAR(50);
-- For ROLE access type: which role is required (DIRECTION, ENGINEERING, etc.)

ALTER TABLE files ADD COLUMN IF NOT EXISTS allowed_roles JSONB DEFAULT '[]'::jsonb;
-- For ROLE access type: multiple roles can be allowed (array of role names)

ALTER TABLE files ADD COLUMN IF NOT EXISTS required_product_id UUID REFERENCES products(id);
-- For PRIVATE access: user must have purchased this specific product

ALTER TABLE files ADD COLUMN IF NOT EXISTS allowed_product_ids JSONB DEFAULT '[]'::jsonb;
-- For PRIVATE access: user must have purchased ANY of these products (array of product IDs)

-- 2. Create file_access_logs table for audit trail
CREATE TABLE IF NOT EXISTS file_access_logs (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL, -- 'VIEW', 'DOWNLOAD', 'ATTEMPTED'
    access_granted BOOLEAN NOT NULL,
    reason TEXT, -- Why access was granted/denied
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_access_logs_file_id ON file_access_logs(file_id);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_user_id ON file_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_created_at ON file_access_logs(created_at);

-- 3. Create file_role_permissions table (alternative to JSONB for roles)
CREATE TABLE IF NOT EXISTS file_role_permissions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE(file_id, role)
);

CREATE INDEX IF NOT EXISTS idx_file_role_permissions_file_id ON file_role_permissions(file_id);
CREATE INDEX IF NOT EXISTS idx_file_role_permissions_role ON file_role_permissions(role);

-- 4. Create file_product_permissions table (alternative to JSONB for products)
CREATE TABLE IF NOT EXISTS file_product_permissions (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE(file_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_file_product_permissions_file_id ON file_product_permissions(file_id);
CREATE INDEX IF NOT EXISTS idx_file_product_permissions_product_id ON file_product_permissions(product_id);

-- 5. Add generated_link_token for public link sharing
ALTER TABLE files ADD COLUMN IF NOT EXISTS public_link_token UUID DEFAULT uuid_generate_v4();
-- Unique token for generating public download links

-- 6. Add public_link_expires_at for time-limited public links
ALTER TABLE files ADD COLUMN IF NOT EXISTS public_link_expires_at TIMESTAMP WITH TIME ZONE;
-- NULL = never expires, timestamp = link expires at this time

-- 7. Add download_count for tracking
ALTER TABLE files ADD COLUMN IF NOT EXISTS download_count INTEGER DEFAULT 0;
-- Track how many times a file has been downloaded

-- 8. Add max_downloads limit (optional)
ALTER TABLE files ADD COLUMN IF NOT EXISTS max_downloads INTEGER;
-- NULL = unlimited downloads, integer = max times file can be downloaded

-- 9. Create function to check file access
CREATE OR REPLACE FUNCTION check_file_access(
    p_file_id UUID,
    p_user_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_file RECORD;
    v_user_role VARCHAR(50);
    v_has_product BOOLEAN;
BEGIN
    -- Get file details
    SELECT * INTO v_file FROM files WHERE id = p_file_id AND is_deleted = FALSE;
    
    IF v_file IS NULL THEN
        RETURN FALSE;
    END IF;
    
    -- File owner always has access
    IF v_file.created_by = p_user_id THEN
        RETURN TRUE;
    END IF;
    
    -- Get user's role
    SELECT role INTO v_user_role FROM user_roles 
    WHERE user_id = p_user_id AND is_active = TRUE 
    ORDER BY hierarchy_level DESC LIMIT 1;
    
    -- PUBLIC access: anyone can download
    IF v_file.access_type = 'PUBLIC' THEN
        -- Check if public link is expired
        IF v_file.public_link_expires_at IS NOT NULL AND v_file.public_link_expires_at < NOW() THEN
            RETURN FALSE;
        END IF;
        
        -- Check max downloads
        IF v_file.max_downloads IS NOT NULL AND v_file.download_count >= v_file.max_downloads THEN
            RETURN FALSE;
        END IF;
        
        RETURN TRUE;
    END IF;
    
    -- ROLE access: check if user has required role
    IF v_file.access_type = 'ROLE' THEN
        -- Check single required role
        IF v_file.required_role IS NOT NULL AND v_user_role = v_file.required_role THEN
            RETURN TRUE;
        END IF;
        
        -- Check allowed_roles array
        IF v_file.allowed_roles IS NOT NULL AND jsonb_array_length(v_file.allowed_roles) > 0 THEN
            IF EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(v_file.allowed_roles) AS role
                WHERE role = v_user_role
            ) THEN
                RETURN TRUE;
            END IF;
        END IF;
        
        -- Check file_role_permissions table
        IF EXISTS (
            SELECT 1 FROM file_role_permissions frp
            WHERE frp.file_id = p_file_id AND frp.role = v_user_role
        ) THEN
            RETURN TRUE;
        END IF;
        
        RETURN FALSE;
    END IF;
    
    -- PRIVATE access: check product purchases
    IF v_file.access_type = 'PRIVATE' OR v_file.access_type IS NULL THEN
        -- Check single required product
        IF v_file.required_product_id IS NOT NULL THEN
            SELECT EXISTS (
                SELECT 1 FROM library
                WHERE user_id = p_user_id 
                AND product_id = v_file.required_product_id
            ) INTO v_has_product;
            
            IF v_has_product THEN
                RETURN TRUE;
            END IF;
        END IF;
        
        -- Check allowed_product_ids array
        IF v_file.allowed_product_ids IS NOT NULL AND jsonb_array_length(v_file.allowed_product_ids) > 0 THEN
            SELECT EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(v_file.allowed_product_ids) AS product_id
                JOIN library l ON l.product_id = product_id::UUID
                WHERE l.user_id = p_user_id
            ) INTO v_has_product;
            
            IF v_has_product THEN
                RETURN TRUE;
            END IF;
        END IF;
        
        -- Check file_product_permissions table
        IF EXISTS (
            SELECT 1 FROM file_product_permissions fpp
            JOIN library l ON l.product_id = fpp.product_id
            WHERE fpp.file_id = p_file_id AND l.user_id = p_user_id
        ) THEN
            RETURN TRUE;
        END IF;
        
        RETURN FALSE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 10. Create function to log file access
CREATE OR REPLACE FUNCTION log_file_access(
    p_file_id UUID,
    p_user_id UUID,
    p_action VARCHAR(20),
    p_access_granted BOOLEAN,
    p_reason TEXT,
    p_ip_address INET,
    p_user_agent TEXT
) RETURNS VOID AS $$
BEGIN
    INSERT INTO file_access_logs (
        file_id, user_id, action, access_granted, reason, ip_address, user_agent
    ) VALUES (
        p_file_id, p_user_id, p_action, p_access_granted, p_reason, p_ip_address, p_user_agent
    );
    
    -- Update download count if access was granted and action is DOWNLOAD
    IF p_access_granted AND p_action = 'DOWNLOAD' THEN
        UPDATE files SET download_count = download_count + 1 WHERE id = p_file_id;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 11. Grant permissions to pixelcraft_user
GRANT ALL ON file_access_logs TO pixelcraft_user;
GRANT ALL ON file_role_permissions TO pixelcraft_user;
GRANT ALL ON file_product_permissions TO pixelcraft_user;
GRANT EXECUTE ON FUNCTION check_file_access TO pixelcraft_user;
GRANT EXECUTE ON FUNCTION log_file_access TO pixelcraft_user;

-- 12. Add comments
COMMENT ON COLUMN files.access_type IS 'PUBLIC (anyone with link), PRIVATE (buyers only), ROLE (specific roles)';
COMMENT ON COLUMN files.required_role IS 'Single role required for ROLE access type';
COMMENT ON COLUMN files.allowed_roles IS 'Array of roles allowed for ROLE access type';
COMMENT ON COLUMN files.required_product_id IS 'Single product purchase required for PRIVATE access';
COMMENT ON COLUMN files.allowed_product_ids IS 'Array of product IDs, user must have purchased ANY of them';
COMMENT ON COLUMN files.public_link_token IS 'Unique token for generating shareable public links';
COMMENT ON COLUMN files.public_link_expires_at IS 'Expiration time for public link (NULL = never expires)';
COMMENT ON COLUMN files.download_count IS 'Number of times this file has been downloaded';
COMMENT ON COLUMN files.max_downloads IS 'Maximum number of downloads allowed (NULL = unlimited)';
COMMENT ON TABLE file_access_logs IS 'Audit trail for all file access attempts';
COMMENT ON TABLE file_role_permissions IS 'Role-based permissions for files (normalized)';
COMMENT ON TABLE file_product_permissions IS 'Product-based permissions for files (normalized)';
