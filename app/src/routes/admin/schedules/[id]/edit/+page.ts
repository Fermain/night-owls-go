import type { PageLoad } from './$types';
import type { ScheduleData } from '$lib/components/admin/schedules/ScheduleForm.svelte';

export const load: PageLoad = async ({ params, fetch }) => {
	const scheduleId = params.id;
	try {
		const response = await fetch(`/api/admin/schedules/${scheduleId}`);
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({ message: 'Failed to load schedule data' }));
			throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
		}
		const schedule = (await response.json()) as ScheduleData;
		return {
			schedule,
			scheduleId
		};
	} catch (error) {
		console.error('Error loading schedule:', error);
		// SvelteKit expects a specific error structure for error pages
		// Or you can return a props field with an error property to handle in the page component
		return {
			schedule: undefined,
			scheduleId,
			error: error instanceof Error ? error.message : 'Unknown error loading schedule'
		};
	}
}; 