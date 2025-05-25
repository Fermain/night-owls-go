import { z } from 'zod';

export const broadcastSchema = z.object({
	broadcast_id: z.number(),
	message: z.string(),
	audience: z.enum(['all', 'admins', 'owls', 'active']),
	sender_user_id: z.number(),
	sender_name: z.string().optional(),
	push_enabled: z.boolean(),
	scheduled_at: z.string().nullable().optional(),
	sent_at: z.string().nullable().optional(),
	status: z.string(),
	recipient_count: z.number(),
	sent_count: z.number(),
	failed_count: z.number(),
	created_at: z.string()
});

export const createBroadcastSchema = z.object({
	message: z.string().min(1, 'Message is required'),
	audience: z.enum(['all', 'admins', 'owls', 'active']),
	push_enabled: z.boolean(),
	scheduled_at: z.string().optional()
});

export type BroadcastData = z.infer<typeof broadcastSchema>;
export type CreateBroadcastData = z.infer<typeof createBroadcastSchema>;
