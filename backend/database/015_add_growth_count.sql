-- Migration to add users_growth_count to analytics snapshot
ALTER TABLE admin_analytics_snapshot ADD COLUMN IF NOT EXISTS users_growth_count INT DEFAULT 0;

-- Ensure last_updated exists (it might be named updated_at in some environments)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin_analytics_snapshot' AND column_name='last_updated') THEN
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin_analytics_snapshot' AND column_name='updated_at') THEN
            ALTER TABLE admin_analytics_snapshot RENAME COLUMN updated_at TO last_updated;
        ELSE
            ALTER TABLE admin_analytics_snapshot ADD COLUMN last_updated TIMESTAMP DEFAULT NOW();
        END IF;
    END IF;
END $$;
