-- Migration: 20230101000002_create_laojun_domains_table
-- Up

CREATE TABLE laojun_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on owner_id for faster lookups
CREATE INDEX idx_laojun_domains_owner_id ON laojun_domains(owner_id);

-- Create index on is_active for filtering
CREATE INDEX idx_laojun_domains_is_active ON laojun_domains(is_active);

-- Create index on name for searching
CREATE INDEX idx_laojun_domains_name ON laojun_domains(name);