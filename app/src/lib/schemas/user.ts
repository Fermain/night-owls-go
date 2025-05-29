import { z } from 'zod';
import type { E164Number } from 'svelte-tel-input/types';
import type { UserRole } from '$lib/types';

// Define possible roles for Zod enum
const userRoles: [UserRole, ...UserRole[]] = ['admin', 'owl', 'guest'];

export const userSchema = z.object({
	// Note: The form uses E164Number | '' for formData.phone,
	// but zod validation here seems to expect a string for the raw input.
	// The transform to E164Number happens before submit.
	// If direct E164Number validation is needed, the schema might need adjustment or a refined type.
	phone: z.string().min(1, 'Phone number is required'), // Or a custom Zod type for E164Number if available
	name: z.string().nullable(),
	role: z.enum(userRoles, { message: 'Role must be admin, owl, or guest' })
});

// This type can be inferred from the schema, but exporting it explicitly can be useful.
export type UserSchemaValues = z.infer<typeof userSchema>;

// This type is used in the form component for its $state.
// It's slightly different from UserSchemaValues because formData.phone can be an empty string initially.
export type UserFormValues = {
	phone: E164Number | ''; // svelte-tel-input uses E164Number or empty string
	name: string | null;
	role: UserRole; // Use UserRole
};

// This type is used as $props for the UserForm and represents existing user data
export type UserData = {
	id: number;
	phone: string; // Assuming phone is always a string for existing users
	name: string | null;
	created_at: string;
	role: UserRole; // Use UserRole
};
