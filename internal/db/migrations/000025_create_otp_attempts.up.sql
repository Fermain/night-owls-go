-- Create OTP attempts table for tracking verification attempts and implementing rate limiting
CREATE TABLE otp_attempts (
    attempt_id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone TEXT NOT NULL,
    attempted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success INTEGER NOT NULL DEFAULT 0, -- 0 = failed, 1 = success
    client_ip TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient lookups by phone and timestamp
CREATE INDEX idx_otp_attempts_phone_time ON otp_attempts(phone, attempted_at);

-- Index for cleanup queries
CREATE INDEX idx_otp_attempts_created_at ON otp_attempts(created_at);

-- Create OTP rate limits table for tracking lockouts
CREATE TABLE otp_rate_limits (
    phone TEXT PRIMARY KEY,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    first_attempt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_attempt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient lockout checks
CREATE INDEX idx_otp_rate_limits_locked_until ON otp_rate_limits(locked_until); 