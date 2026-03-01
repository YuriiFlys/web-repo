-- Drop old string assignee column if it exists.
-- Run this against your project-management database once after deploying the new schema.
ALTER TABLE tasks DROP COLUMN IF EXISTS assignee;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS description TEXT;
