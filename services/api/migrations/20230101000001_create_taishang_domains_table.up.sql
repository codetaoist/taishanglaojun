-- Migration: 20230101000001_create_taishang_domains_table
-- Up

CREATE TABLE taishang_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on owner_id for faster lookups
CREATE INDEX idx_taishang_domains_owner_id ON taishang_domains(owner_id);

-- Create index on is_active for filtering
CREATE INDEX idx_taishang_domains_is_active ON taishang_domains(is_active);

-- Create index on name for searching
CREATE INDEX idx_taishang_domains_name ON taishang_domains(name);