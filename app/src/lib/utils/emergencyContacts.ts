import type { components } from '$lib/types/api';

/**
 * Emergency contact types from OpenAPI spec
 */
export type EmergencyContact = components['schemas']['api.EmergencyContactResponse'];
export type CreateEmergencyContactRequest =
	components['schemas']['api.CreateEmergencyContactRequest'];
export type UpdateEmergencyContactRequest =
	components['schemas']['api.UpdateEmergencyContactRequest'];

/**
 * Sort emergency contacts by display order
 */
export function sortContactsByDisplayOrder(contacts: EmergencyContact[]): EmergencyContact[] {
	return [...contacts].sort((a, b) => (a.display_order ?? 0) - (b.display_order ?? 0));
}

/**
 * Filter emergency contacts based on search term
 */
export function filterContacts(
	contacts: EmergencyContact[],
	searchTerm: string
): EmergencyContact[] {
	if (!searchTerm) return sortContactsByDisplayOrder(contacts);

	const term = searchTerm.toLowerCase();
	return contacts
		.filter(
			(contact) =>
				contact.name?.toLowerCase().includes(term) ||
				contact.number?.includes(term) ||
				contact.description?.toLowerCase().includes(term)
		)
		.sort((a, b) => (a.display_order ?? 0) - (b.display_order ?? 0));
}

/**
 * Format phone number for display (basic formatting)
 */
export function formatPhoneNumber(number: string): string {
	// Remove all non-digit characters
	const cleaned = number.replace(/\D/g, '');

	// Basic South African number formatting
	if (cleaned.length === 10 && cleaned.startsWith('0')) {
		// Format: 012 345 6789
		return `${cleaned.slice(0, 3)} ${cleaned.slice(3, 6)} ${cleaned.slice(6)}`;
	}

	if (cleaned.length === 9 && !cleaned.startsWith('0')) {
		// Format: 12 345 6789 (without leading 0)
		return `${cleaned.slice(0, 2)} ${cleaned.slice(2, 5)} ${cleaned.slice(5)}`;
	}

	// Return original if no pattern matches
	return number;
}

/**
 * Validate emergency contact form data
 */
export function validateEmergencyContact(data: Partial<CreateEmergencyContactRequest>): {
	isValid: boolean;
	errors: string[];
} {
	const errors: string[] = [];

	if (!data.name?.trim()) {
		errors.push('Name is required');
	}

	if (!data.number?.trim()) {
		errors.push('Phone number is required');
	} else if (data.number.trim().length < 9) {
		errors.push('Phone number must be at least 9 digits');
	}

	if (data.display_order !== undefined && data.display_order < 1) {
		errors.push('Display order must be at least 1');
	}

	return {
		isValid: errors.length === 0,
		errors
	};
}

/**
 * Get the default emergency contact from a list
 */
export function getDefaultContact(contacts: EmergencyContact[]): EmergencyContact | undefined {
	return contacts.find((contact) => contact.is_default);
}

/**
 * Check if a contact can be deleted (not the default contact)
 */
export function canDeleteContact(contact: EmergencyContact): boolean {
	return !contact.is_default;
}
