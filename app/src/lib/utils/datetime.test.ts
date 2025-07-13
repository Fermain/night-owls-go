import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getTimeUntil } from './datetime';

describe('getTimeUntil', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('should return "in X minutes" for future times', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		const futureTime = '2024-01-01T10:30:00Z';
		expect(getTimeUntil(futureTime)).toBe('in 30 minutes');
	});

	it('should return "in about X hours" for times over 60 minutes away', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		const futureTime = '2024-01-01T12:00:00Z';
		expect(getTimeUntil(futureTime)).toBe('in about 2 hours');
	});

	it('should return "less than a minute ago" for current time', () => {
		const now = new Date('2024-01-01T10:00:00Z');
		vi.setSystemTime(now);

		expect(getTimeUntil(now.toISOString())).toBe('less than a minute ago');
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
		expect(getTimeUntil('2024-01-01T10:59:00Z')).toBe('in about 1 hour');

		// Exactly 1 hour
		expect(getTimeUntil('2024-01-01T11:00:00Z')).toBe('in about 1 hour');

		// Just over 1 hour
		expect(getTimeUntil('2024-01-01T11:01:00Z')).toBe('in about 1 hour');
	});
});
