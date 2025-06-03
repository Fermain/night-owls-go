import { authenticatedFetch } from '$lib/utils/api';
import type { BroadcastData, CreateBroadcastData } from '$lib/schemas/broadcast';

export async function getBroadcasts(): Promise<BroadcastData[]> {
	const response = await authenticatedFetch('/api/admin/broadcasts');
	if (!response.ok) {
		throw new Error(`Failed to fetch broadcasts: ${response.status}`);
	}
	return response.json();
}

export async function createBroadcast(data: CreateBroadcastData): Promise<BroadcastData> {
	const response = await authenticatedFetch('/api/admin/broadcasts', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});
	if (!response.ok) {
		throw new Error(`Failed to create broadcast: ${response.status}`);
	}
	return response.json();
}

export async function getBroadcast(id: number): Promise<BroadcastData> {
	const response = await authenticatedFetch(`/api/admin/broadcasts/${id}`);
	if (!response.ok) {
		throw new Error(`Failed to fetch broadcast: ${response.status}`);
	}
	return response.json();
}

export async function deleteBroadcast(id: number): Promise<{ message: string }> {
	const response = await authenticatedFetch(`/api/admin/broadcasts/${id}`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error(`Failed to delete broadcast: ${response.status}`);
	}
	return response.json();
}
