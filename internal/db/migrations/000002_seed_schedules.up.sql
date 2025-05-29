-- Seed initial schedules for 2025

-- Daily Evening Patrol - 6 PM shifts
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Daily Evening Patrol',
    '0 18 * * *', -- Every day at 6 PM
    '2025-01-01',
    '2025-12-31',
    120,
    'Africa/Johannesburg'
);

-- Weekend Morning Watch - Saturday and Sunday at 6 AM and 10 AM
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Weekend Morning Watch',
    '0 6,10 * * 6,0', -- Saturday and Sunday at 6 AM and 10 AM
    '2025-01-01',
    '2025-12-31',
    240,
    'Africa/Johannesburg'
);

-- Weekday Lunch Security - Monday to Friday at noon
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Weekday Lunch Security',
    '0 12 * * 1-5', -- Monday to Friday at noon
    '2025-01-01',
    '2025-12-31',
    60,
    'Africa/Johannesburg'
); 