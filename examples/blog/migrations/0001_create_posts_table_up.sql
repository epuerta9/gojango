-- Migration: Create posts table
-- Created: 2024-01-01 12:00:00
-- Description: Initial table for blog posts

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author_email TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index for better query performance
CREATE INDEX idx_posts_published ON posts(published);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- Insert some sample data
INSERT INTO posts (title, content, author_email, published) VALUES 
('Welcome to Gojango Database Layer', 
 'This post demonstrates the database integration capabilities of Gojango. With support for multiple database drivers, migration management, and Ent ORM integration, building data-driven applications is straightforward.',
 'admin@example.com',
 TRUE),

('Getting Started with Migrations',
 'Database migrations in Gojango are managed through a robust system that tracks schema changes over time. You can create, apply, and rollback migrations using the CLI tools.',
 'developer@example.com', 
 TRUE),

('Advanced Database Features',
 'Gojango provides advanced database features including connection pooling, transaction management, and integration with the Ent ORM for type-safe database operations.',
 'expert@example.com',
 FALSE);

-- Create users table for future use
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    full_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users
INSERT INTO users (email, username, full_name) VALUES
('admin@example.com', 'admin', 'Administrator'),
('developer@example.com', 'dev', 'Developer User'),
('expert@example.com', 'expert', 'Database Expert');