import { type Page } from '@playwright/test';

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

export const mockShifts = [
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

export async function setupApiMocks(page: Page) {
	// Mock ping endpoint for MSW testing
	await page.route('**/api/ping', async (route) => {
		await route.fulfill({
			status: 501,
			contentType: 'application/json',
			body: JSON.stringify({
				message: 'MSW intercepted - ping endpoint mocked',
				intercepted: true
			})
		});
	});

	// Mock broadcasts endpoint
	await page.route('**/api/broadcasts**', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{
						id: 1,
						message: 'Community safety reminder: Please report suspicious activity',
						audience: 'all_users',
						recipient_count: 42,
						status: 'sent',
						push_enabled: true,
						created_at: new Date().toISOString()
					},
					{
						id: 2,
						message: 'Patrol schedule updated for this weekend',
						audience: 'owls_only',
						recipient_count: 15,
						status: 'pending',
						push_enabled: false,
						created_at: new Date(Date.now() - 86400000).toISOString()
					}
				])
			});
		} else if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData() || '{}');
			await route.fulfill({
				status: 201,
				contentType: 'application/json',
				body: JSON.stringify({
					id: Date.now(),
					...body,
					status: 'sent',
					created_at: new Date().toISOString()
				})
			});
		}
	});

	// Mock admin dashboard endpoint
	await page.route('**/api/admin/dashboard**', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				shifts: {
					total_upcoming: 12,
					unassigned: 3,
					assigned: 9
				},
				users: {
					total: 45,
					active_this_month: 23,
					new_this_week: 2
				},
				recent_activity: [
					{
						id: 1,
						type: 'shift_booking',
						description: 'John Doe booked Morning Patrol',
						timestamp: new Date().toISOString()
					},
					{
						id: 2,
						type: 'user_registered',
						description: 'New user Jane Smith registered',
						timestamp: new Date(Date.now() - 3600000).toISOString()
					}
				]
			})
		});
	});

	// Mock admin schedules all-slots endpoint
	await page.route('**/api/admin/schedules/all-slots**', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify([
				{
					id: 1,
					schedule_id: 1,
					schedule_name: 'Morning Patrol',
					start_time: '2024-12-25T08:00:00Z',
					end_time: '2024-12-25T12:00:00Z',
					is_assigned: false,
					assigned_user_name: null,
					buddy_name: null
				},
				{
					id: 2,
					schedule_id: 2,
					schedule_name: 'Evening Watch',
					start_time: '2024-12-25T18:00:00Z',
					end_time: '2024-12-25T22:00:00Z',
					is_assigned: true,
					assigned_user_name: 'John Doe',
					buddy_name: 'Jane Smith'
				}
			])
		});
	});

	// Mock emergency contacts endpoint
	await page.route('**/api/emergency-contacts', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{
						id: 1,
						name: 'Emergency Services',
						phone: '112',
						type: 'emergency'
					},
					{
						id: 2,
						name: 'Local Police',
						phone: '10111',
						type: 'police'
					},
					{
						id: 3,
						name: 'Medical Emergency',
						phone: '999',
						type: 'medical'
					}
				])
			});
		} else if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData() || '{}');
			await route.fulfill({
				status: 201,
				contentType: 'application/json',
				body: JSON.stringify({
					id: Date.now(),
					...body,
					created_at: new Date().toISOString()
				})
			});
		}
	});

	// Mock authentication endpoints
	await page.route('**/api/auth/register', async (route) => {
		const request = route.request();
		const body = JSON.parse(request.postData() || '{}');

		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				success: true,
				message: 'Registration successful!',
				user: {
					id: Date.now(),
					name: body.name || 'User',
					phone: body.phone || '',
					role: 'guest'
				}
			})
		});
	});

	await page.route('**/api/auth/verify', async (route) => {
		const request = route.request();
		const body = JSON.parse(request.postData() || '{}');

		// Accept any 6-digit OTP for testing
		const code = body.code || body.otp || '';
		if (!/^\d{6}$/.test(code)) {
			await route.fulfill({
				status: 400,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'Invalid OTP format' })
			});
			return;
		}

		const user = Object.values(mockUsers).find((u) => u.phone === body.phone) || {
			id: Date.now(),
			name: 'Test User',
			phone: body.phone || '',
			role: 'guest'
		};

		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				success: true,
				message: 'Login successful!',
				token: 'mock-jwt-token',
				user
			})
		});
	});

	// Mock schedules endpoints
	await page.route('**/api/admin/schedules', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockSchedules)
			});
		} else if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData() || '{}');
			const newSchedule = {
				id: Date.now(),
				...body,
				created_at: new Date().toISOString()
			};

			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(newSchedule)
			});
		}
	});

	// Mock users endpoints
	await page.route('**/api/admin/users', async (route) => {
		if (route.request().method() === 'GET') {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(Object.values(mockUsers))
			});
		} else if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData() || '{}');
			const newUser = {
				id: Date.now(),
				...body,
				created_at: new Date().toISOString()
			};

			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(newUser)
			});
		}
	});

	// Mock shifts endpoints
	await page.route('**/shifts/available', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockShifts)
		});
	});

	await page.route('**/bookings', async (route) => {
		if (route.request().method() === 'POST') {
			const body = JSON.parse(route.request().postData() || '{}');

			const booking = {
				id: Date.now(),
				schedule_id: body.schedule_id,
				user_id: mockUsers.volunteer.id,
				user_name: mockUsers.volunteer.name,
				buddy_name: body.buddy_name,
				status: 'booked',
				created_at: new Date().toISOString()
			};

			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					success: true,
					message: 'Shift booked successfully!',
					booking
				})
			});
		}
	});

	// Mock dashboard endpoints
	await page.route('**/api/admin/dashboard/shifts', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify([])
		});
	});

	console.log('✅ API mocks configured for page');
}
