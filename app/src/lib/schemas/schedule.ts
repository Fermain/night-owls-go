import cronstrue from 'cronstrue';
import * as z from 'zod';

export type ZodSchemaValues = z.infer<typeof scheduleZodSchema>;

export const scheduleZodSchema = z
	.object({
		name: z.string().min(1, 'Schedule name is required'),
		cron_expr: z
			.string()
			.min(1, 'CRON expression is required')
			.refine(
				(val) => {
					try {
						cronstrue.toString(val);
						return true;
					} catch (_e) {
						console.error('Invalid cron expression:', _e);
						return false;
					}
				},
				{ message: 'Invalid CRON expression format' }
			),
		start_date: z.date().nullable().optional(),
		end_date: z.date().nullable().optional()
	})
	.refine(
		(data) => {
			if (data.start_date && data.end_date && data.start_date > data.end_date) {
				return false;
			}
			return true;
		},
		{
			message: 'End date cannot be before start date',
			path: ['end_date']
		}
	);
