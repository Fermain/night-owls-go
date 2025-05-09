-- Seed initial schedules

-- Summer Patrol (Nov 2024 - Apr 2025)
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Summer Patrol (Nov-Apr)',
    '0 0,2 * 11-12,1-4 6,0,1', -- Sat/Sun/Mon, 00:00 & 02:00, Nov-Apr (Months: Nov,Dec,Jan,Feb,Mar,Apr; DaysOfWeek: Sat,Sun,Mon)
    '2024-11-01',
    '2025-04-30',
    120,
    'Africa/Johannesburg' -- Example timezone, adjust if necessary or make it configurable
);

-- Winter Patrol (May 2025 - Oct 2025)
INSERT INTO schedules (name, cron_expr, start_date, end_date, duration_minutes, timezone)
VALUES (
    'Winter Patrol (May-Oct)',
    '0 1,3 * 5-10 6,0,1', -- Sat/Sun/Mon, 01:00 & 03:00, May-Oct (Months: May,Jun,Jul,Aug,Sep,Oct; DaysOfWeek: Sat,Sun,Mon)
    '2025-05-01',
    '2025-10-31',
    120,
    'Africa/Johannesburg' -- Example timezone
); 