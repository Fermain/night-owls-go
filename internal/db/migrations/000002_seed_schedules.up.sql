-- Seed initial schedules

-- Old schedule - daily midnight shifts (with existing bookings and reports)
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Old schedule',
    '0 0 * * *', -- Every day at midnight
    '2025-01-01',
    '2025-12-31',
    120,
    'Africa/Johannesburg'
);

-- New schedule - daily midnight shifts (fresh start)
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'New schedule',
    '0 0 * * *', -- Every day at midnight
    '2025-01-01',
    '2025-12-31',
    120,
    'Africa/Johannesburg'
); 