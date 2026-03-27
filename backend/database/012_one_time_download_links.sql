-- Migration: One-Time Download Links
-- Description: Adds support for generating unique one-time download links for private files
-- Date: 2026-02-28

-- 1. Create one_time_download_tokens table
CREATE TABLE IF NOT EXISTS one_time_download_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,
    download_count INTEGER DEFAULT 0 NOT NULL,
    max_downloads INTEGER DEFAULT 1 NOT NULL,
    ip_address INET,
    user_agent TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_one_time_tokens_token ON one_time_download_tokens(token);
CREATE INDEX IF NOT EXISTS idx_one_time_tokens_file_id ON one_time_download_tokens(file_id);
CREATE INDEX IF NOT EXISTS idx_one_time_tokens_user_id ON one_time_download_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_one_time_tokens_expires_at ON one_time_download_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_one_time_tokens_is_used ON one_time_download_tokens(is_used);

-- 2. Create function to clean up expired tokens
CREATE OR REPLACE FUNCTION cleanup_expired_download_tokens() RETURNS INTEGER AS $$
DECLARE
    v_deleted_count INTEGER;
BEGIN
    DELETE FROM one_time_download_tokens
    WHERE expires_at < NOW() OR (is_used = TRUE AND used_at < NOW() - INTERVAL '24 hours');
    
    GET DIAGNOSTICS v_deleted_count = ROW_COUNT;
    RETURN v_deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 3. Create function to validate and use token
CREATE OR REPLACE FUNCTION validate_and_use_download_token(
    p_token UUID,
    p_ip_address INET,
    p_user_agent TEXT
) RETURNS TABLE(
    file_id UUID,
    is_valid BOOLEAN,
    error_message TEXT
) AS $$
DECLARE
    v_token_record RECORD;
    v_error TEXT;
BEGIN
    -- Get token record
    SELECT * INTO v_token_record FROM one_time_download_tokens
    WHERE token = p_token;

    -- Check if token exists
    IF v_token_record IS NULL THEN
        RETURN QUERY SELECT NULL::UUID, FALSE, 'Invalid or expired token';
        RETURN;
    END IF;

    -- Check if already used
    IF v_token_record.is_used = TRUE THEN
        RETURN QUERY SELECT v_token_record.file_id, FALSE, 'Token has already been used';
        RETURN;
    END IF;

    -- Check if expired
    IF v_token_record.expires_at < NOW() THEN
        RETURN QUERY SELECT v_token_record.file_id, FALSE, 'Token has expired';
        RETURN;
    END IF;

    -- Check max downloads
    IF v_token_record.max_downloads IS NOT NULL AND v_token_record.download_count >= v_token_record.max_downloads THEN
        RETURN QUERY SELECT v_token_record.file_id, FALSE, 'Download limit reached for this token';
        RETURN;
    END IF;

    -- Update token usage
    UPDATE one_time_download_tokens
    SET 
        download_count = download_count + 1,
        is_used = (download_count + 1 >= max_downloads),
        used_at = CASE 
            WHEN download_count + 1 >= max_downloads THEN NOW()
            ELSE used_at
        END,
        ip_address = COALESCE(p_ip_address, ip_address),
        user_agent = COALESCE(p_user_agent, user_agent)
    WHERE id = v_token_record.id;

    -- Return success with file_id
    RETURN QUERY SELECT v_token_record.file_id, TRUE, ''::TEXT;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 4. Grant permissions
GRANT ALL ON one_time_download_tokens TO pixelcraft_user;
GRANT EXECUTE ON FUNCTION cleanup_expired_download_tokens TO pixelcraft_user;
GRANT EXECUTE ON FUNCTION validate_and_use_download_token TO pixelcraft_user;

-- 5. Add comments
COMMENT ON TABLE one_time_download_tokens IS 'Stores one-time use download tokens for private files';
COMMENT ON COLUMN one_time_download_tokens.token IS 'Unique token used in download URL';
COMMENT ON COLUMN one_time_download_tokens.expires_at IS 'Token expiration time';
COMMENT ON COLUMN one_time_download_tokens.used_at IS 'When token was first used';
COMMENT ON COLUMN one_time_download_tokens.is_used IS 'Whether token has been fully consumed';
COMMENT ON COLUMN one_time_download_tokens.max_downloads IS 'Maximum downloads allowed per token (default: 1)';
COMMENT ON FUNCTION cleanup_expired_download_tokens IS 'Removes expired and used tokens older than 24 hours';
COMMENT ON FUNCTION validate_and_use_download_token IS 'Validates token and increments usage counter, returns validity status';
