-- Create users table
CREATE TABLE IF NOT EXISTS lao_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS lao_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES lao_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON lao_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON lao_sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON lao_sessions(expires_at);

-- Create default admin user (password: admin123)
INSERT INTO lao_users (username, email, password_hash, role)
VALUES (
    'admin',
    'admin@taishanglaojun.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi.', -- password: admin123
    'admin'
)
ON CONFLICT (username) DO NOTHING;