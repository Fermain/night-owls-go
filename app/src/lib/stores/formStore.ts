import { persisted } from 'svelte-persisted-store';

export interface FormData {
	lastPhoneNumber: string | null;
	lastName: string | null;
}

const initialFormData: FormData = {
	lastPhoneNumber: null,
	lastName: null
};

// Create persisted store for form data
export const formStore = persisted<FormData>('form-data', initialFormData);

// Helper functions
export function saveUserData(phoneNumber: string, name?: string) {
	formStore.update((data) => ({
		...data,
		lastPhoneNumber: phoneNumber,
		lastName: name?.trim() || data.lastName
	}));
}

export function clearUserData() {
	formStore.set(initialFormData);
}
