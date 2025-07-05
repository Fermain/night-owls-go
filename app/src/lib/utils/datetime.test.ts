import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getTimeUntil } from './datetime';

describe('getTimeUntil', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('should return "in Xm" for future times', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		const futureTime = '2024-01-01T10:30:00Z';
		expect(getTimeUntil(futureTime)).toBe('in 30m');
	});

	it('should return "in Xh Ym" for times over 60 minutes away', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		const futureTime = '2024-01-01T12:00:00Z';
		expect(getTimeUntil(futureTime)).toBe('in 2h 0m');
	});

	it('should return "in 0m" for current time', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		expect(getTimeUntil(now.toISOString())).toBe('in 0m');
	});

	it('should return "Started" for past times', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		const pastTime = '2024-01-01T09:30:00Z';
		expect(getTimeUntil(pastTime)).toBe('Started');
	});

	it('should handle edge cases near time boundaries', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		// Just under 1 hour
		expect(getTimeUntil('2024-01-01T10:59:00Z')).toBe('in 59m');

		// Exactly 1 hour
		expect(getTimeUntil('2024-01-01T11:00:00Z')).toBe('in 1h 0m');

		// Just over 1 hour
		expect(getTimeUntil('2024-01-01T11:01:00Z')).toBe('in 1h 1m');
	});
});
