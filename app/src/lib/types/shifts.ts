export interface Shift {
	schedule_id: number;
	schedule_name: string;
	start_time: string;
	end_time: string;
	timezone?: string;
	is_booked: boolean;
}

export interface ProcessedShift extends Shift {
	is_tonight: boolean;
	priority: string;
	slots_available: number;
	total_slots: number;
}

export type FilterOption = 'all' | 'tonight' | 'available' | 'urgent';
