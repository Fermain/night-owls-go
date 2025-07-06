CREATE TABLE calendar_tokens (
    token_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE, -- SHA-256 hash of the actual token
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    last_accessed_at DATETIME,
    access_count INTEGER DEFAULT 0,
    is_revoked BOOLEAN DEFAULT 0,
    
    -- Indexes for performance
    UNIQUE(user_id, token_hash)
);

-- Index for efficient token lookup
CREATE INDEX idx_calendar_tokens_hash ON calendar_tokens(token_hash) WHERE is_revoked = 0;

-- Index for cleanup of expired tokens
CREATE INDEX idx_calendar_tokens_expires ON calendar_tokens(expires_at);

-- Index for user token management
CREATE INDEX idx_calendar_tokens_user ON calendar_tokens(user_id, is_revoked);