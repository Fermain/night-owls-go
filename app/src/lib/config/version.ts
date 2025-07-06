// Application version automatically injected from package.json at build time
export const APP_VERSION = __APP_VERSION__;

// Application metadata
export const APP_INFO = {
	name: 'Night Owls',
	fullName: 'Mount Moreland Community Watch Platform',
	version: APP_VERSION,
	organization: 'Mount Moreland Community Watch'
} as const;
