// app/src/lib/types/index.ts
// This file will hold shared TypeScript type and interface declarations for the application.

export type AdminShiftSlot = {
	schedule_id: number;
	schedule_name: string;
	start_time: string; // ISO date string
	end_time: string; // ISO date string
	timezone?: string | null;
	is_booked: boolean;
	booking_id?: number | null;
	user_name?: string | null;
	user_phone?: string | null;
};

export interface UserSessionData {
	isAuthenticated: boolean;
	id: string | null;
	name: string | null;
	phone: string | null;
	role: UserRole | null;
	token: string | null;
}

// export {}; // Remove initial export {} if types are added

// New types being added:
export type SQLNullString = {
	String: string;
	Valid: boolean;
};

export type SQLNullTime = {
	Time: string; // Assuming string representation of time, adjust if it's Date
	Valid: boolean;
};

export type Schedule = {
	schedule_id: number;
	name: string;
	cron_expr: string;
	timezone: string | null; // From schema, was SQLNullString
	start_date: string | null; // From schema, originally SQLNullTime, now string to match form/api
	end_date: string | null; // From schema, originally SQLNullTime, now string to match form/api
	duration_minutes: number; // Duration of shifts in minutes
	is_active: boolean;
	created_at: string;
	updated_at: string;
	next_run_time?: string | null; // Added, often useful
	slot_count?: number; // Added, often useful
};

export type UserRole = 'admin' | 'owl' | 'guest';
