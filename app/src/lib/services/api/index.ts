// API Services
export { SchedulesApiService } from './schedules';
export { UsersApiService } from './users';
export { ShiftsApiService } from './shifts';
export { ReportsApiService } from './reports';

// Import for convenience object
import { SchedulesApiService } from './schedules';
import { UsersApiService } from './users';
import { ShiftsApiService } from './shifts';
import { ReportsApiService } from './reports';

// Re-export for convenience
export const ApiServices = {
	schedules: SchedulesApiService,
	users: UsersApiService,
	shifts: ShiftsApiService,
	reports: ReportsApiService
} as const;
