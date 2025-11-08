CREATE TABLE IF NOT EXISTS lao_token_blacklist (
    id SERIAL PRIMARY KEY, 
    token_hash VARCHAR(255) NOT NULL UNIQUE, 
    user_id INTEGER NOT NULL, 
    reason VARCHAR(255), 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(), 
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL, 
    FOREIGN KEY (user_id) REFERENCES lao_users(id) ON DELETE CASCADE
);