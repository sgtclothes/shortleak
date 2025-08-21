-- init-db.sql
-- This file will be executed when PostgreSQL container starts

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create any additional extensions you might need
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- You can add any initial database setup here
-- CREATE TABLE IF NOT EXISTS example_table (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT NOW()
-- );