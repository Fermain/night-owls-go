-- Drop OTP rate limiting tables (reverse migration)
DROP INDEX IF EXISTS idx_otp_rate_limits_locked_until;
DROP TABLE IF EXISTS otp_rate_limits;

DROP INDEX IF EXISTS idx_otp_attempts_created_at;
DROP INDEX IF EXISTS idx_otp_attempts_phone_time;
DROP TABLE IF EXISTS otp_attempts; 