// API Services
export { SchedulesApiService } from './schedules';
export { UsersApiService } from './users';
export { ShiftsApiService } from './shifts';
export { ReportsApiService } from './reports';
export { BookingsApiService } from './bookings';

// Import for convenience object
import { SchedulesApiService } from './schedules';
import { UsersApiService } from './users';
import { ShiftsApiService } from './shifts';
import { ReportsApiService } from './reports';
import { BookingsApiService } from './bookings';

// Re-export for convenience
export const ApiServices = {
	schedules: SchedulesApiService,
	users: UsersApiService,
	shifts: ShiftsApiService,
	reports: ReportsApiService,
	bookings: BookingsApiService
} as const;
