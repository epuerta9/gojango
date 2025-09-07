-- Rollback for: Create posts table
-- Created: 2024-01-01 12:00:00
-- Description: Rollback initial table creation

-- Drop indexes first
DROP INDEX IF EXISTS idx_posts_published;
DROP INDEX IF EXISTS idx_posts_created_at;

-- Drop tables
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS users;