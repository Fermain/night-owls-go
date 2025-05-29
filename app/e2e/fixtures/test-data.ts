export const testUsers = {
	admin: {
		name: 'Test Admin',
		phone: '+27821111111',
		role: 'admin'
	},
	volunteer: {
		name: 'Test Volunteer',
		phone: '+27821111112',
		role: 'owl'
	},
	guest: {
		name: 'Test Guest',
		phone: '+27821111113',
		role: 'guest'
	}
} as const;

export const testSchedules = {
	morningPatrol: {
		name: 'Morning Patrol',
		description: 'Early morning Night Owls Control patrol',
		cronExpression: '0 6 * * *',
		duration: 120,
		positions: 2,
		timezone: 'Africa/Johannesburg'
	},
	eveningWatch: {
		name: 'Evening Watch',
		description: 'Evening monitoring shift',
		cronExpression: '0 18 * * 1-5',
		duration: 180,
		positions: 1,
		timezone: 'Africa/Johannesburg'
	},
	weekendShift: {
		name: 'Weekend Shift',
		description: 'Weekend community patrol',
		cronExpression: '0 10 * * 6,0',
		duration: 240,
		positions: 3,
		timezone: 'Africa/Johannesburg'
	}
} as const;

export const testValidationCases = {
	schedules: {
		validCronExpressions: [
			'0 9 * * *', // Every day at 9 AM
			'0 18 * * 1-5', // Weekdays at 6 PM
			'0 10 * * 6,0', // Weekends at 10 AM
			'30 14 * * *' // Every day at 2:30 PM
		],
		invalidCronExpressions: [
			'invalid',
			'60 9 * * *', // Invalid minute
			'0 25 * * *', // Invalid hour
			'0 9 * * 8' // Invalid day of week
		],
		validDurations: [60, 120, 180, 240, 300],
		invalidDurations: [0, -30, 1440, 2000],
		validPositions: [1, 2, 3, 5, 10],
		invalidPositions: [0, -1, 100]
	}
} as const;

export const testOTPs = {
	valid: '123456',
	invalid: 'abc123',
	tooShort: '123',
	tooLong: '1234567'
} as const;

export const testPhoneNumbers = {
	valid: ['+27821234567', '+1234567890', '+447911123456'],
	invalid: ['123', 'abc', '+27821', 'not-a-phone']
} as const;

// Utility function to generate unique test data
export function generateUniqueTestData() {
	const timestamp = Date.now().toString().slice(-6);
	const random = Math.floor(Math.random() * 1000)
		.toString()
		.padStart(3, '0');
	const uniqueId = timestamp + random.slice(-3); // Combine timestamp + random for uniqueness

	return {
		user: {
			name: `Test User ${uniqueId}`,
			phone: `+27821${uniqueId}`,
			role: 'guest' as const
		},
		schedule: {
			name: `Test Schedule ${uniqueId}`,
			description: `Generated test schedule ${uniqueId}`,
			cronExpression: '0 12 * * *',
			duration: 120,
			positions: 2,
			timezone: 'Africa/Johannesburg'
		}
	};
}

// Common test scenarios
export const testScenarios = {
	authentication: {
		newUserRegistration: {
			name: testUsers.guest.name,
			phone: testUsers.guest.phone,
			expectedRole: 'guest'
		},
		existingUserLogin: {
			phone: testUsers.volunteer.phone,
			expectedRole: 'owl'
		},
		adminLogin: {
			phone: testUsers.admin.phone,
			expectedRole: 'admin'
		}
	},

	scheduleManagement: {
		createBasicSchedule: testSchedules.morningPatrol,
		createComplexSchedule: testSchedules.weekendShift,
		invalidScheduleData: {
			name: '',
			description: 'Invalid schedule',
			cronExpression: 'invalid-cron',
			duration: -1,
			positions: 0,
			timezone: 'Invalid/Timezone'
		}
	},

	shiftBooking: {
		soloBooking: {
			shiftId: 1,
			buddyName: undefined
		},
		buddyBooking: {
			shiftId: 1,
			buddyName: 'Buddy Partner'
		},
		conflictingBooking: {
			shiftId: 2, // Assume this shift is already full
			expectedError: 'Shift is full'
		}
	}
} as const;
