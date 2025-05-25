import { http, HttpResponse } from 'msw';

// Type definitions
interface CreateScheduleRequest {
  name: string;
  description: string;
  cron_expression: string;
  duration_minutes: number;
  timezone: string;
  positions_available: number;
}

interface UpdateScheduleRequest {
  name?: string;
  description?: string;
  cron_expression?: string;
  duration_minutes?: number;
  timezone?: string;
  positions_available?: number;
}

interface CreateUserRequest {
  name: string;
  phone: string;
  role?: string;
}

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
    id: 1,
    schedule_id: 1,
    start_time: '2024-12-22T18:00:00Z',
    end_time: '2024-12-22T20:00:00Z',
    positions_available: 2,
    positions_filled: 0,
    bookings: []
  },
  {
    id: 2,
    schedule_id: 2,
    start_time: '2024-12-21T10:00:00Z',
    end_time: '2024-12-21T13:00:00Z',
    positions_available: 1,
    positions_filled: 1,
    bookings: [{
      id: 1,
      user_id: 2,
      user_name: 'Bob Volunteer',
      status: 'booked'
    }]
  }
];

// API handlers
export const handlers = [
  // Authentication endpoints
  http.post('/api/auth/register', async ({ request }) => {
    const body = await request.json() as { name: string; phone: string };
    return HttpResponse.json({
      success: true,
      message: 'Registration successful!',
      user: {
        id: Date.now(),
        name: body.name,
        phone: body.phone,
        role: 'guest'
      }
    });
  }),

  http.post('/api/auth/verify', async ({ request }) => {
    const body = await request.json() as { phone: string; otp: string };
    
    // Accept any 6-digit OTP for testing
    if (!/^\d{6}$/.test(body.otp)) {
      return HttpResponse.json(
        { error: 'Invalid OTP format' },
        { status: 400 }
      );
    }

    const user = Object.values(mockUsers).find(u => u.phone === body.phone) || {
      id: Date.now(),
      name: 'Test User',
      phone: body.phone,
      role: 'guest'
    };

    return HttpResponse.json({
      success: true,
      message: 'Login successful!',
      token: 'mock-jwt-token',
      user
    });
  }),

  http.post('/api/auth/dev-login', async ({ request }) => {
    const body = await request.json() as { phone: string };
    const user = Object.values(mockUsers).find(u => u.phone === body.phone);
    
    if (!user) {
      return HttpResponse.json(
        { error: 'User not found' },
        { status: 404 }
      );
    }

    return HttpResponse.json({
      success: true,
      token: 'mock-jwt-token',
      user
    });
  }),

  // Schedules endpoints
  http.get('/api/admin/schedules', () => {
    return HttpResponse.json(mockSchedules);
  }),

  http.post('/api/admin/schedules', async ({ request }) => {
    const body = await request.json() as CreateScheduleRequest;
    const newSchedule = {
      id: Date.now(),
      ...body,
      created_at: new Date().toISOString()
    };
    mockSchedules.push(newSchedule);
    return HttpResponse.json(newSchedule);
  }),

  http.get('/api/admin/schedules/:id', ({ params }) => {
    const schedule = mockSchedules.find(s => s.id === Number(params.id));
    if (!schedule) {
      return HttpResponse.json(
        { error: 'Schedule not found' },
        { status: 404 }
      );
    }
    return HttpResponse.json(schedule);
  }),

  http.put('/api/admin/schedules/:id', async ({ params, request }) => {
    const body = await request.json() as UpdateScheduleRequest;
    const index = mockSchedules.findIndex(s => s.id === Number(params.id));
    if (index === -1) {
      return HttpResponse.json(
        { error: 'Schedule not found' },
        { status: 404 }
      );
    }
    mockSchedules[index] = { ...mockSchedules[index], ...body };
    return HttpResponse.json(mockSchedules[index]);
  }),

  http.delete('/api/admin/schedules/:id', ({ params }) => {
    const index = mockSchedules.findIndex(s => s.id === Number(params.id));
    if (index === -1) {
      return HttpResponse.json(
        { error: 'Schedule not found' },
        { status: 404 }
      );
    }
    mockSchedules.splice(index, 1);
    return HttpResponse.json({ success: true });
  }),

  // Users endpoints
  http.get('/api/admin/users', () => {
    return HttpResponse.json(Object.values(mockUsers));
  }),

  http.post('/api/admin/users', async ({ request }) => {
    const body = await request.json() as CreateUserRequest;
    const newUser = {
      id: Date.now(),
      ...body,
      created_at: new Date().toISOString()
    };
    return HttpResponse.json(newUser);
  }),

  // Shifts endpoints
  http.get('/shifts/available', () => {
    return HttpResponse.json(mockShifts.filter(s => s.positions_filled < s.positions_available));
  }),

  http.post('/bookings', async ({ request }) => {
    const body = await request.json() as { shift_id: number; buddy_name?: string };
    const shift = mockShifts.find(s => s.id === body.shift_id);
    
    if (!shift) {
      return HttpResponse.json(
        { error: 'Shift not found' },
        { status: 404 }
      );
    }
    
    if (shift.positions_filled >= shift.positions_available) {
      return HttpResponse.json(
        { error: 'Shift is full' },
        { status: 400 }
      );
    }

    const booking = {
      id: Date.now(),
      shift_id: body.shift_id,
      user_id: mockUsers.volunteer.id,
      user_name: mockUsers.volunteer.name,
      buddy_name: body.buddy_name,
      status: 'booked',
      created_at: new Date().toISOString()
    };

    shift.bookings.push(booking);
    shift.positions_filled += 1;

    return HttpResponse.json({
      success: true,
      message: 'Shift booked successfully!',
      booking
    });
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