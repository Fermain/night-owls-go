import { http, HttpResponse } from 'msw';

// Mock data fixtures
export const mockUsers = {
	admin: {
		id: 1,
		name: 'Alice Admin',
		phone: '+27821234567',
		role: 'admin'
	},
	volunteer: {
		id: 2,
		name: 'Bob Volunteer',
		phone: '+27821234568',
		role: 'owl'
	},
	guest: {
		id: 3,
		name: 'Charlie Guest',
		phone: '+27821234569',
		role: 'guest'
	}
};

export const mockSchedules = [
	{
		id: 1,
		name: 'Evening Patrol',
		description: 'Evening community patrol',
		cron_expression: '0 18 * * *',
		duration_minutes: 120,
		timezone: 'Africa/Johannesburg',
		positions_available: 2
	},
	{
		id: 2,
		name: 'Weekend Watch',
		description: 'Weekend monitoring',
		cron_expression: '0 10 * * 6,0',
		duration_minutes: 180,
		timezone: 'Africa/Johannesburg',
		positions_available: 1
	}
];

// API handlers
export const handlers = [
	// Authentication endpoints
	http.post('/api/auth/register', async ({ request }) => {
		const body = await request.json();
		return HttpResponse.json({
			success: true,
			message: 'Registration successful!',
			user: {
				id: Date.now(),
				name: body?.name || 'User',
				phone: body?.phone || '',
				role: 'guest'
			}
		});
	}),

	http.post('/api/auth/verify', async ({ request }) => {
		const body = await request.json();

		// Accept any 6-digit OTP for testing
		const code = body?.code || body?.otp || '';
		if (!/^\d{6}$/.test(code)) {
			return HttpResponse.json({ error: 'Invalid OTP format' }, { status: 400 });
		}

		const user = Object.values(mockUsers).find((u) => u.phone === body?.phone) || {
			id: Date.now(),
			name: 'Test User',
			phone: body?.phone || '',
			role: 'guest'
		};

		return HttpResponse.json({
			success: true,
			message: 'Login successful!',
			token: 'mock-jwt-token',
			user
		});
	}),

	// Schedules endpoints
	http.get('/api/admin/schedules', () => {
		return HttpResponse.json(mockSchedules);
	}),

	http.post('/api/admin/schedules', async ({ request }) => {
		const body = await request.json();
		const newSchedule = {
			id: Date.now(),
			...body,
			created_at: new Date().toISOString()
		};
		mockSchedules.push(newSchedule);
		return HttpResponse.json(newSchedule);
	}),

	// Users endpoints
	http.get('/api/admin/users', () => {
		return HttpResponse.json(Object.values(mockUsers));
	}),

	http.post('/api/admin/users', async ({ request }) => {
		const body = await request.json();
		const newUser = {
			id: Date.now(),
			...body,
			created_at: new Date().toISOString()
		};
		return HttpResponse.json(newUser);
	}),

	// Shifts endpoints
	http.get('/shifts/available', () => {
		// Return data in the format expected by AvailableShiftSlot interface
		const availableShifts = [
			{
				schedule_id: 1,
				schedule_name: 'Evening Patrol',
				start_time: '2024-12-22T18:00:00Z',
				end_time: '2024-12-22T20:00:00Z',
				timezone: 'Africa/Johannesburg',
				is_booked: false
			},
			{
				schedule_id: 2,
				schedule_name: 'Weekend Watch',
				start_time: '2024-12-21T10:00:00Z',
				end_time: '2024-12-21T13:00:00Z',
				timezone: 'Africa/Johannesburg',
				is_booked: true
			},
			{
				schedule_id: 1,
				schedule_name: 'Evening Patrol',
				start_time: '2024-12-23T18:00:00Z',
				end_time: '2024-12-23T20:00:00Z',
				timezone: 'Africa/Johannesburg',
				is_booked: false
			}
		];
		return HttpResponse.json(availableShifts);
	}),

	http.post('/bookings', async ({ request }) => {
		const body = await request.json();

		const booking = {
			id: Date.now(),
			schedule_id: body.schedule_id,
			user_id: mockUsers.volunteer.id,
			user_name: mockUsers.volunteer.name,
			buddy_name: body.buddy_name,
			status: 'booked',
			created_at: new Date().toISOString()
		};

		return HttpResponse.json({
			success: true,
			message: 'Shift booked successfully!',
			booking
		});
	}),

	// Dashboard data endpoints
	http.get('/api/admin/dashboard/shifts', () => {
		return HttpResponse.json([]);
	}),

	// Catch-all for unmocked endpoints
	http.get('*', ({ request }) => {
		console.warn(`Unmocked GET request to ${request.url}`);
		return HttpResponse.json({ error: 'Endpoint not mocked' }, { status: 501 });
	}),

	http.post('*', ({ request }) => {
		console.warn(`Unmocked POST request to ${request.url}`);
		return HttpResponse.json({ error: 'Endpoint not mocked' }, { status: 501 });
	})
];
