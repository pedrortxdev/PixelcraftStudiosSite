-- Migration 008: Update check_file_access function to use new relational tables
-- Updates the function to use file_permission_roles and file_permission_products

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

        -- Check allowed_roles array (JSON - legacy support)
        IF v_file.allowed_roles IS NOT NULL AND jsonb_array_length(v_file.allowed_roles) > 0 THEN
            IF EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(v_file.allowed_roles) AS role
                WHERE role = v_user_role
            ) THEN
                RETURN TRUE;
            END IF;
        END IF;

        -- Check file_permission_roles table (NEW - primary source)
        IF EXISTS (
            SELECT 1 FROM file_permission_roles fpr
            WHERE fpr.file_id = p_file_id AND fpr.role_name = v_user_role
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
                SELECT 1 FROM user_purchases
                WHERE user_id = p_user_id
                AND product_id = v_file.required_product_id
            ) INTO v_has_product;

            IF v_has_product THEN
                RETURN TRUE;
            END IF;
        END IF;

        -- Check allowed_product_ids array (JSON - legacy support)
        IF v_file.allowed_product_ids IS NOT NULL AND jsonb_array_length(v_file.allowed_product_ids) > 0 THEN
            SELECT EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(v_file.allowed_product_ids) AS product_id
                JOIN user_purchases up ON up.product_id = product_id::UUID
                WHERE up.user_id = p_user_id
            ) INTO v_has_product;

            IF v_has_product THEN
                RETURN TRUE;
            END IF;
        END IF;

        -- Check file_permission_products table (NEW - primary source)
        IF EXISTS (
            SELECT 1 FROM file_permission_products fpp
            JOIN user_purchases up ON up.product_id = fpp.product_id
            WHERE fpp.file_id = p_file_id AND up.user_id = p_user_id
        ) THEN
            RETURN TRUE;
        END IF;

        RETURN FALSE;
    END IF;

    RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Grant permissions
GRANT EXECUTE ON FUNCTION check_file_access TO pixelcraft_user;
