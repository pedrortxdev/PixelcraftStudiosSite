-- Migration: Add Support Attachments
-- Description: Add attachment_url and attachment_type to support_messages

DO $$ BEGIN
    ALTER TABLE support_messages ADD COLUMN attachment_url TEXT;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

DO $$ BEGIN
    ALTER TABLE support_messages ADD COLUMN attachment_type VARCHAR(50);
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;
