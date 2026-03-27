-- Migration: Add image_url to plans table
ALTER TABLE plans ADD COLUMN IF NOT EXISTS image_url TEXT;
